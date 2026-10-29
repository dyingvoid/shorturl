package infrastructure_kafka

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/IBM/sarama"
	"github.com/dyingvoid/urlshort/url/internal/domain"
)

type UserViewProducer struct {
	producer sarama.AsyncProducer
	cfg      Config
	log      *slog.Logger
}

func NewUserViewProducer(log *slog.Logger, cfg Config) (*UserViewProducer, error) {
	saramaCfg := sarama.NewConfig()

	saramaCfg.ChannelBufferSize = 512

	saramaCfg.Producer.Flush.Frequency = 10 * time.Millisecond
	saramaCfg.Producer.Flush.Messages = 5

	saramaCfg.Producer.RequiredAcks = sarama.WaitForLocal
	saramaCfg.Producer.Retry.Max = 1

	saramaCfg.Producer.Return.Errors = true
	saramaCfg.Producer.Return.Successes = true

	producer, err := sarama.NewAsyncProducer(cfg.Brokers, saramaCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create async producer: %w", err)
	}

	go func() {
		for err := range producer.Errors() {
			log.Error("producer:", "err", err)
		}
	}()

	go func() {
		for msg := range producer.Successes() {
			log.Info(
				"message produced",
				"topic", msg.Topic,
				"partition", msg.Partition,
				"offset", msg.Offset,
			)
		}
	}()

	return &UserViewProducer{
		producer: producer,
		cfg:      cfg,
		log:      log,
	}, nil
}

func (p *UserViewProducer) ProduceUserView(view domain.UserView) error {
	payload, err := json.Marshal(view)
	if err != nil {
		return fmt.Errorf("marshal user view: %w", err)
	}

	p.producer.Input() <- &sarama.ProducerMessage{
		Topic: p.cfg.Topic,
		Value: sarama.ByteEncoder(payload),
	}

	return nil
}

func (p *UserViewProducer) Close() error {
	if err := p.producer.Close(); err != nil {
		return fmt.Errorf("close kafka producer: %w", err)
	}

	return nil
}
