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
}

func NewUserViewProducer(log *slog.Logger, cfg Config) (*UserViewProducer, error) {
	saramaCfg := sarama.NewConfig()

	saramaCfg.ChannelBufferSize = 512

	saramaCfg.Producer.Flush.Frequency = 10 * time.Millisecond
	saramaCfg.Producer.Flush.Messages = 500

	saramaCfg.Producer.RequiredAcks = sarama.WaitForLocal
	saramaCfg.Producer.Retry.Max = 1

	saramaCfg.Producer.Return.Errors = true

	producer, err := sarama.NewAsyncProducer(cfg.Brokers, saramaCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create async producer: %w", err)
	}

	go func() {
		for err := range producer.Errors() {
			log.Error("producer:", "err", err)
		}
	}()

	return &UserViewProducer{
		producer: producer,
		cfg:      cfg,
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
