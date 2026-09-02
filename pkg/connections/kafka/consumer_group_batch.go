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
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.uber.org/zap"

	"demo_bot/pkg/telemetry"
)

// ConsumerGroupBatchOpt allows to customize consumer group batch.
type ConsumerGroupBatchOpt func(cg *ConsumerGroupBatch)

func WithCustomErrorHandlerConsumerGroupBatchOpt(eh func(err error)) ConsumerGroupBatchOpt {
	return func(cg *ConsumerGroupBatch) {
		cg.errHandler = eh
	}
}

func WithNotAckOnErrorConsumerGroupBatchOpt() ConsumerGroupBatchOpt {
	return func(cg *ConsumerGroupBatch) {
		cg.ackOnError = false
	}
}

func WithBufferSizeConsumerGroupBatchOpt(bufferSize int) ConsumerGroupBatchOpt {
	return func(cg *ConsumerGroupBatch) {
		cg.bufferSize = bufferSize
	}
}

type pconsumerbatch struct {
	name  string
	group string

	topic     string
	partition int32

	logger *zap.Logger

	kTracer *kotel.Tracer

	ctx           context.Context
	contextCancel context.CancelFunc
	done          chan struct{}
	recs          chan kgo.FetchTopicPartition

	ackOnError bool
	callback   ConsumerBatchCallback

	markRecordsFunc func(...*kgo.Record)
	errHandler      func(error)
	backoff         backoff.BackOff
}

type ConsumerGroupBatch struct {
	kf     *Kafka
	logger *zap.Logger

	group string

	opts []kgo.Opt

	ackOnError bool
	bufferSize int
	errHandler func(error)

	client *kgo.Client

	handlers  map[string]ConsumerBatchCallback
	consumers map[tp]*pconsumerbatch

	wg wait.Group
}

func NewConsumerGroupBatch(
	kf *Kafka,
	group string,
	opts ...ConsumerGroupBatchOpt,
) (*ConsumerGroupBatch, error) {
	group = kf.formatWithPrefix(group)

	cg := &ConsumerGroupBatch{
		kf:     kf,
		logger: kf.logger.Named("consumer_group").With(zap.String("group", group)),

		group: group,
		opts:  make([]kgo.Opt, 0, len(opts)+len(kf.opts)+6),

		ackOnError: true, // ack if was error
		bufferSize: 20,

		handlers:  make(map[string]ConsumerBatchCallback, 1),
		consumers: make(map[tp]*pconsumerbatch),

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

func (cg *ConsumerGroupBatch) Subscribe(
	topic string,
	callback ConsumerBatchCallback,
) {
	topic = cg.kf.formatWithPrefix(topic)

	cg.handlers[topic] = callback
	cg.client.AddConsumeTopics(topic)
}

func (cg *ConsumerGroupBatch) Close() {
	cg.client.Close()
	cg.wg.Wait()
}

//nolint:dupl
func (cg *ConsumerGroupBatch) Consume(maxPollRecords int) {
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
					consumerGroupBatchPollTotal.WithLabelValues(
						cg.kf.name,
						cg.group,
						p.Topic,
						strconv.FormatInt(int64(p.Partition), 10),
						strconv.FormatBool(false),
					).Add(float64(len(p.Records)))
				default:
					consumerGroupBatchPollTotal.WithLabelValues(
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

func (cg *ConsumerGroupBatch) Ping(ctx context.Context) error {
	if err := cg.client.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping Kafka: %w", err)
	}

	return nil
}

func (cg *ConsumerGroupBatch) assigned(
	_ context.Context,
	cl *kgo.Client,
	assigned map[string][]int32,
) {
	for topic, partitions := range assigned {
		for _, partition := range partitions {
			ctx, cancel := context.WithCancel(context.Background()) //nolint:gosec
			pc := &pconsumerbatch{
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

func (cg *ConsumerGroupBatch) revoked(
	ctx context.Context,
	cl *kgo.Client,
	revoked map[string][]int32,
) {
	cg.killConsumers(revoked)
	if err := cl.CommitMarkedOffsets(ctx); err != nil {
		cg.logger.Warn("revoke commit, failed", zap.Error(err))
	}
}

func (cg *ConsumerGroupBatch) lost(_ context.Context, _ *kgo.Client, lost map[string][]int32) {
	cg.killConsumers(lost)
}

func (cg *ConsumerGroupBatch) killConsumers(lost map[string][]int32) {
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

func (pc *pconsumerbatch) consume() {
	defer close(pc.done)
	pc.logger.Info("starting batch",
		zap.String("topic", pc.topic),
		zap.Int32("partition", pc.partition),
	)
	defer pc.logger.Info("stop batch",
		zap.String("topic", pc.topic),
		zap.Int32("partition", pc.partition),
	)

	const optimisticPollSize = 100
	payloads := make([]Payload, 0, optimisticPollSize)

	for {
		select {
		case <-pc.ctx.Done():
			return
		case p := <-pc.recs:
			select {
			case <-pc.ctx.Done():
				return
			default:
			}

			start := time.Now()
			lastIndex := len(p.Records) - 1

			ctx, span := pc.kTracer.WithProcessSpan(p.Records[0])
			span.SetAttributes([]attribute.KeyValue{
				semconv.MessagingKafkaConsumerGroupKey.String(pc.group),
				attribute.Key("messaging.kafka.message.begin_offset").Int(int(p.Records[0].Offset)),
				attribute.Key("messaging.kafka.message.end_offset").Int(int(p.Records[lastIndex].Offset)),
			}...)

			payloads = payloads[:0]
			for _, record := range p.Records {
				payloads = append(payloads, convertRecordToPayload(record))
			}
			err := pc.callback(ctx, payloads)

			consumerGroupBatchHandle.WithLabelValues(
				pc.name, pc.group, pc.topic,
				telemetry.ErrLabelValue(err),
			).Observe(time.Since(start).Seconds())

			if err != nil {
				pc.errHandler(err)
				span.RecordError(err)

				if pc.ackOnError {
					pc.markRecordsFunc(p.Records[lastIndex])
				}
			} else {
				pc.markRecordsFunc(p.Records[lastIndex])
			}

			span.End()
		}
	}
}
