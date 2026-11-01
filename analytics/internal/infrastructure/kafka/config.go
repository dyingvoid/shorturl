package kafka

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Brokers            []string      `envconfig:"KAFKA_BROKERS" default:"localhost:9092"`
	Topic              string        `envconfig:"KAFKA_TOPIC" default:"click_events"`
	GroupID            string        `envconfig:"KAFKA_GROUP_ID" default:"analytics"`
	BatchSize          int           `envconfig:"BATCH_SIZE" default:"100"`
	BatchFlushInterval time.Duration `envconfig:"BATCH_FLUSH_INTERVAL" default:"5s"`
}

func New() (Config, error) {
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		return Config{}, fmt.Errorf("process envconfig: %w", err)
	}

	switch {
	case config.BatchSize <= 0:
		return Config{}, fmt.Errorf("BATCH_SIZE must be positive, got %d", config.BatchSize)
	case config.BatchFlushInterval <= 0:
		return Config{}, fmt.Errorf("BATCH_FLUSH_INTERVAL must be positive, got %s", config.BatchFlushInterval)
	}

	return config, nil
}

func NewMust() Config {
	config, err := New()
	if err != nil {
		err = fmt.Errorf("get config: %w", err)
		panic(err)
	}

	return config
}
