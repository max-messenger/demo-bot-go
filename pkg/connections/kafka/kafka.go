package kafka

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/plain"
	"github.com/twmb/franz-go/pkg/sasl/scram"
	"github.com/twmb/franz-go/plugin/kotel"
	"github.com/twmb/franz-go/plugin/kzap"
	"go.uber.org/zap"
)

type Key []byte
type Payload []byte

type ConsumerCallback func(ctx context.Context, payload Payload) error
type ConsumerBatchCallback func(ctx context.Context, payload []Payload) error

const defaultClientName = "kafka"

// Kafka contains all things for default connection, metrics, and logger.
type Kafka struct {
	logger *zap.Logger
	cfg    *Config

	name string

	kTracer *kotel.Tracer

	opts []kgo.Opt
}

// nolint:cyclop
func NewKafka(
	logger *zap.Logger,
	cfg *Config,
	name string,
) (*Kafka, error) {
	if cfg == nil {
		return nil, fmt.Errorf("empty config")
	}
	if name == "" {
		name = defaultClientName
	}

	opts := []kgo.Opt{
		kgo.SeedBrokers(cfg.Seeds...),
		kgo.WithLogger(kzap.New(logger)),
	}

	logger.Info("kafka mechanism config", zap.String("mechanism", string(cfg.SASLMechanism)))

	switch cfg.SASLMechanism {
	case SASLMechanismNone:
		// do nothing
	case SASLMechanismPlain:
		opts = append(opts, kgo.SASL(plain.Auth{
			User: cfg.Username,
			Pass: cfg.Password,
		}.AsMechanism()))
	case SASLMechanismScramSHA256:
		opts = append(opts, kgo.SASL(scram.Auth{
			User: cfg.Username,
			Pass: cfg.Password,
		}.AsSha256Mechanism()))
	case SASLMechanismScramSHA512:
		opts = append(opts, kgo.SASL(scram.Auth{
			User: cfg.Username,
			Pass: cfg.Password,
		}.AsSha512Mechanism()))
	default:
		return nil, fmt.Errorf("unknown sasl mechanism: %s", cfg.SASLMechanism)
	}

	if cfg.TLS {
		opts = append(opts, kgo.DialTLS())
	}
	if cfg.Certificate != "" {
		caCert, err := os.ReadFile(cfg.Certificate)
		if err != nil {
			return nil, fmt.Errorf("read certificate: %w", err)
		}

		caCertPool := x509.NewCertPool()
		ok := caCertPool.AppendCertsFromPEM(caCert)
		if !ok {
			return nil, fmt.Errorf("append certs from pem")
		}

		opts = append(opts, kgo.DialTLSConfig(
			&tls.Config{
				MinVersion: tls.VersionTLS12,
				RootCAs:    caCertPool,
			}))
	}

	if cfg.AllowAutoTopicCreation {
		opts = append(opts, kgo.AllowAutoTopicCreation())
	}

	kTracer := kotel.NewTracer()
	kotelOps := []kotel.Opt{
		kotel.WithTracer(kTracer),
	}
	kotelService := kotel.NewKotel(kotelOps...)
	opts = append(opts, kgo.WithHooks(kotelService.Hooks()...))

	return &Kafka{
		logger:  logger,
		kTracer: kTracer,
		name:    name,
		cfg:     cfg,
		opts:    opts,
	}, nil
}

func (k *Kafka) formatWithPrefix(str string) string {
	if k.cfg.Prefix != "" {
		return fmt.Sprintf("%s_%s", k.cfg.Prefix, str)
	}

	return str
}

func (k *Kafka) producerOpts() []kgo.Opt {
	res := make([]kgo.Opt, 0)
	pCfg := k.cfg.Producer

	if pCfg.DisableIdempotentWrite {
		res = append(res, kgo.DisableIdempotentWrite())
	}
	if pCfg.RequestTimeout > 0 {
		res = append(res, kgo.ProduceRequestTimeout(pCfg.RequestTimeout))
	}
	if pCfg.RecordRetries > 0 {
		res = append(res, kgo.RecordRetries(pCfg.RecordRetries))
	}
	if pCfg.MaxRecordBatchBytes > 0 {
		res = append(res, kgo.ProducerBatchMaxBytes(pCfg.MaxRecordBatchBytes))
	}
	if pCfg.MaxBufferedRecords > 0 {
		res = append(res, kgo.MaxBufferedRecords(pCfg.MaxBufferedRecords))
	}
	if pCfg.MaxBufferedBytes > 0 {
		res = append(res, kgo.MaxBufferedBytes(pCfg.MaxBufferedBytes))
	}

	return res

}
func convertPayloadToRecord(
	topic string,
	key Key,
	payload Payload,
) *kgo.Record {
	return &kgo.Record{
		Topic: topic,
		Key:   key,
		Value: payload,
	}
}

func convertPayloadsToRecords(
	topic string,
	payload ...Payload,
) []*kgo.Record {
	res := make([]*kgo.Record, 0, len(payload))
	for _, p := range payload {
		res = append(res, convertPayloadToRecord(topic, nil, p))
	}

	return res
}

func convertRecordToPayload(r *kgo.Record) Payload {
	return r.Value
}
