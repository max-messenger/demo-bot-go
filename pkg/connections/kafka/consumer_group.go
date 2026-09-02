package kafka

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/chapsuk/wait"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/plugin/kotel"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.uber.org/zap"

	"demo_bot/pkg/telemetry"
)

// ConsumerGroupOpt allows to customize consumer group.
type ConsumerGroupOpt func(cg *ConsumerGroup)

func WithCustomErrorHandlerConsumerGroupOpt(eh func(err error)) ConsumerGroupOpt {
	return func(cg *ConsumerGroup) {
		cg.errHandler = eh
	}
}

func WithNotAckOnErrorConsumerGroupOpt() ConsumerGroupOpt {
	return func(cg *ConsumerGroup) {
		cg.ackOnError = false
	}
}

func WithBufferSize(bufferSize int) ConsumerGroupOpt {
	return func(cg *ConsumerGroup) {
		cg.bufferSize = bufferSize
	}
}

type tp struct {
	t string
	p int32
}

type pconsumer struct {
	group string
	name  string

	topic     string
	partition int32

	logger *zap.Logger

	kTracer *kotel.Tracer

	ctx           context.Context
	contextCancel context.CancelFunc
	done          chan struct{}
	recs          chan kgo.FetchTopicPartition

	ackOnError bool
	callback   ConsumerCallback

	markRecordsFunc func(...*kgo.Record)
	errHandler      func(error)
	backoff         backoff.BackOff
}

type ConsumerGroup struct {
	kf     *Kafka
	logger *zap.Logger

	group string

	opts []kgo.Opt

	ackOnError bool
	bufferSize int
	errHandler func(error)

	client *kgo.Client

	handlers  map[string]ConsumerCallback
	consumers map[tp]*pconsumer

	wg wait.Group
}

func NewConsumerGroup(
	kf *Kafka,
	group string,
	opts ...ConsumerGroupOpt,
) (*ConsumerGroup, error) {
	group = kf.formatWithPrefix(group)

	cg := &ConsumerGroup{
		kf:     kf,
		logger: kf.logger.Named("consumer_group").With(zap.String("group", group)),

		group: group,
		opts:  make([]kgo.Opt, 0, len(opts)+len(kf.opts)+6),

		ackOnError: true, // ack if was error
		bufferSize: 20,

		handlers:  make(map[string]ConsumerCallback, 1),
		consumers: make(map[tp]*pconsumer),

		wg: wait.Group{},
	}

	cg.errHandler = func(err error) {
		cg.logger.Warn("Kafka: consumer error", zap.Error(err))
	}

	for _, opt := range opts {
		opt(cg)
	}

	mainOpts := []kgo.Opt{
		kgo.ConsumerGroup(group),
		kgo.OnPartitionsAssigned(cg.assigned),
		kgo.OnPartitionsRevoked(cg.revoked),
		kgo.OnPartitionsLost(cg.lost),
		kgo.AutoCommitMarks(),
		kgo.BlockRebalanceOnPoll(),
	}

	cg.opts = append(cg.opts, kf.opts...)
	cg.opts = append(cg.opts, mainOpts...)

	cl, err := kgo.NewClient(cg.opts...)
	if err != nil {
		return nil, err
	}

	cg.client = cl

	return cg, nil
}

func (cg *ConsumerGroup) Subscribe(
	topic string,
	callback ConsumerCallback,
) {
	topic = cg.kf.formatWithPrefix(topic)

	cg.handlers[topic] = callback
	cg.client.AddConsumeTopics(topic)
}

func (cg *ConsumerGroup) Close() {
	cg.client.Close()
	cg.wg.Wait()
}

//nolint:dupl
func (cg *ConsumerGroup) Consume(maxPollRecords int) {
	cg.wg.Add(func() {
		for {
			fetches := cg.client.PollRecords(context.Background(), maxPollRecords)
			if fetches.IsClientClosed() {
				cg.logger.Info("stop consumer group")

				return
			}
			fetches.EachError(func(t string, p int32, err error) {
				cg.errHandler(fmt.Errorf("topic: %s, partition: %d error: %w", t, p, err))
			})

			fetches.EachPartition(func(p kgo.FetchTopicPartition) {
				if len(p.Records) == 0 {
					return
				}

				tpKey := tp{p.Topic, p.Partition}
				cons, ok := cg.consumers[tpKey]
				if !ok {
					cg.errHandler(
						fmt.Errorf(
							"unknown topic partition in consumer_group: %s:%d",
							p.Topic,
							p.Partition,
						),
					)

					return
				}

				select {
				case cons.recs <- p:
					consumerGroupPollTotal.WithLabelValues(
						cg.kf.name,
						cg.group,
						p.Topic,
						strconv.FormatInt(int64(p.Partition), 10),
						strconv.FormatBool(false),
					).Add(float64(len(p.Records)))
				default:
					consumerGroupPollTotal.WithLabelValues(
						cg.kf.name,
						cg.group,
						p.Topic,
						strconv.FormatInt(int64(p.Partition), 10),
						strconv.FormatBool(true),
					).Add(float64(len(p.Records)))

					topic, partition := p.Topic, p.Partition

					cg.client.PauseFetchPartitions(map[string][]int32{
						topic: {partition},
					})

					go func() {
						defer func() {
							cg.client.ResumeFetchPartitions(map[string][]int32{
								topic: {partition},
							})
						}()

						select {
						case <-cons.ctx.Done():
						case cons.recs <- p:
						}
					}()
				}
			})

			cg.client.AllowRebalance()
		}
	})
}

func (cg *ConsumerGroup) Ping(ctx context.Context) error {
	if err := cg.client.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping Kafka: %w", err)
	}

	return nil
}

func (cg *ConsumerGroup) assigned(_ context.Context, cl *kgo.Client, assigned map[string][]int32) {
	for topic, partitions := range assigned {
		for _, partition := range partitions {
			ctx, cancel := context.WithCancel(context.Background()) //nolint:gosec
			pc := &pconsumer{
				name:       cg.kf.name,
				logger:     cg.logger.Named("topic_partition_consumer"),
				group:      cg.group,
				topic:      topic,
				partition:  partition,
				ackOnError: cg.ackOnError,

				kTracer:  cg.kf.kTracer,
				callback: cg.handlers[topic],

				ctx:           ctx,
				contextCancel: cancel,
				done:          make(chan struct{}),
				recs:          make(chan kgo.FetchTopicPartition, cg.bufferSize),

				markRecordsFunc: func(r ...*kgo.Record) {
					cl.MarkCommitRecords(r...)
				},
				errHandler: cg.errHandler,
				backoff:    cg.kf.cfg.Consumer.ConfigBackoff.GetBackOff(),
			}
			cg.consumers[tp{topic, partition}] = pc
			cg.wg.Add(pc.consume)
		}
	}
}

func (cg *ConsumerGroup) revoked(ctx context.Context, cl *kgo.Client, revoked map[string][]int32) {
	cg.killConsumers(revoked)
	if err := cl.CommitMarkedOffsets(ctx); err != nil {
		cg.logger.Warn("revoke commit, failed", zap.Error(err))
	}
}

func (cg *ConsumerGroup) lost(_ context.Context, _ *kgo.Client, lost map[string][]int32) {
	cg.killConsumers(lost)
}

func (cg *ConsumerGroup) killConsumers(lost map[string][]int32) {
	var wg sync.WaitGroup
	defer wg.Wait()

	for topic, partitions := range lost {
		for _, partition := range partitions {
			tpKey := tp{topic, partition}
			pc := cg.consumers[tpKey]
			delete(cg.consumers, tpKey)
			pc.contextCancel()
			pc.logger.Info(
				"waiting for work to finish",
				zap.String("topic", topic),
				zap.Int32("partition", partition),
			)
			wg.Go(func() { <-pc.done })
		}
	}
}

func (pc *pconsumer) consume() {
	defer close(pc.done)
	pc.logger.Info("starting", zap.String("topic", pc.topic), zap.Int32("partition", pc.partition))
	defer pc.logger.Info("stop", zap.String("topic", pc.topic), zap.Int32("partition", pc.partition))

	for {
		select {
		case <-pc.ctx.Done():
			return
		case p := <-pc.recs:
			for _, r := range p.Records {
				select {
				case <-pc.ctx.Done():
					return
				default:
				}

				start := time.Now()

				ctx, span := pc.kTracer.WithProcessSpan(r)
				span.SetAttributes(semconv.MessagingKafkaConsumerGroupKey.String(pc.group))

				var err error
				tryNum := 0

				err = backoff.RetryNotify(
					func() error {
						tryNum++

						return pc.callback(ctx, convertRecordToPayload(r))
					},
					backoff.WithContext(pc.backoff, pc.ctx),
					func(err error, duration time.Duration) {
						pc.logger.Warn(
							"consumer call callback, retrying",
							zap.String("topic", pc.topic),
							zap.Int32("partition", pc.partition),
							zap.Int("try_num", tryNum),
							zap.Duration("retry_delay", duration),
							zap.Error(err),
						)
					},
				)

				consumerGroupHandle.WithLabelValues(
					pc.name, pc.group, pc.topic,
					telemetry.ErrLabelValue(err),
				).Observe(time.Since(start).Seconds())

				if err != nil {
					pc.errHandler(err)
					span.RecordError(err)

					if pc.ackOnError {
						pc.markRecordsFunc(r)
					}
				} else {
					pc.markRecordsFunc(r)
				}

				span.End()
			}
		}
	}
}
