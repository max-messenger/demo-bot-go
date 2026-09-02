package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"demo_bot/pkg/telemetry"
)

type ProducerOpt func(*Producer)

type Producer struct {
	kf *Kafka
	cl *kgo.Client
}

func NewProducer(kf *Kafka) (*Producer, error) {
	p := &Producer{
		kf: kf,
	}

	cl, err := kgo.NewClient(append(kf.opts, kf.producerOpts()...)...)
	if err != nil {
		return nil, err
	}

	p.cl = cl

	return p, nil

}

func (p *Producer) Send(ctx context.Context, topic string, payload Payload) error {
	topic = p.kf.formatWithPrefix(topic)

	ctx, span := otel.Tracer("kafka").Start(ctx, "producer_send", trace.WithAttributes(
		attribute.String("kafka.topic", topic),
		attribute.String("kafka.operation", "send"),
		attribute.String("kafka.name", p.kf.name),
	))
	defer span.End()

	err := p.cl.ProduceSync(
		ctx,
		convertPayloadToRecord(topic, nil, payload),
	).FirstErr()

	producerTotal.WithLabelValues(p.kf.name, topic, telemetry.ErrLabelValue(err)).Inc()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return fmt.Errorf("produce: %w", err)
	}

	return nil
}

func (p *Producer) SendWithKey(
	ctx context.Context,
	topic string,
	key Key,
	payload Payload,
) error {
	topic = p.kf.formatWithPrefix(topic)

	ctx, span := otel.Tracer("kafka").Start(ctx, "producer_send_with_key", trace.WithAttributes(
		attribute.String("kafka.topic", topic),
		attribute.String("kafka.operation", "send"),
		attribute.String("kafka.name", p.kf.name),
	))
	defer span.End()

	err := p.cl.ProduceSync(
		ctx,
		convertPayloadToRecord(topic, key, payload),
	).FirstErr()

	producerTotal.WithLabelValues(p.kf.name, topic, telemetry.ErrLabelValue(err)).Inc()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		return fmt.Errorf("produce: %w", err)
	}

	return nil
}

func (p *Producer) SendBatch(ctx context.Context, topic string, payloads ...Payload) []error {
	topic = p.kf.formatWithPrefix(topic)

	ctx, span := otel.Tracer("kafka").Start(ctx, "producer_send_batch", trace.WithAttributes(
		attribute.String("kafka.topic", topic),
		attribute.String("kafka.operation", "send"),
		attribute.String("kafka.name", p.kf.name),
	))
	defer span.End()

	results := p.cl.ProduceSync(ctx, convertPayloadsToRecords(topic, payloads...)...)
	errors := make([]error, 0, len(payloads))
	for i := range results {
		r := &results[i]

		producerTotal.WithLabelValues(p.kf.name, topic, telemetry.ErrLabelValue(r.Err)).Inc()

		if r.Err != nil {
			span.RecordError(r.Err)
			errors = append(errors, fmt.Errorf("produce error: %w", r.Err))

			continue
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

func (p *Producer) SendAsync(ctx context.Context, topic string, payload Payload) {
	start := time.Now()
	topic = p.kf.formatWithPrefix(topic)

	ctx, span := otel.Tracer("kafka").Start(ctx, "producer_send_async", trace.WithAttributes(
		attribute.String("kafka.topic", topic),
		attribute.String("kafka.operation", "send"),
		attribute.String("kafka.name", p.kf.name),
	))
	defer span.End()

	producerAsyncMessagesInFlight.WithLabelValues(p.kf.name, topic).Add(1)

	p.cl.Produce(ctx,
		convertPayloadToRecord(topic, nil, payload),
		func(_ *kgo.Record, err error) {
			if err != nil {
				span.RecordError(err)
				p.kf.logger.Error("Kafka: async producer error", zap.Error(err), zap.String("topic", topic))
			}

			producerTotal.WithLabelValues(p.kf.name, topic, telemetry.ErrLabelValue(err)).Inc()

			producerAsyncHandle.WithLabelValues(
				p.kf.name, topic, telemetry.ErrLabelValue(err)).
				Observe(time.Since(start).Seconds())
			producerAsyncMessagesInFlight.WithLabelValues(p.kf.name, topic).Add(-1)
		})
}

func (p *Producer) Ping(ctx context.Context) error {
	if err := p.cl.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping Kafka: %w", err)
	}

	return nil
}

func (p *Producer) Close(_ context.Context) error {
	p.cl.Close()

	return nil
}
