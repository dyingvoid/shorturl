package infrastructure_kafka

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Brokers []string `envconfig:"KAFKA_BROKERS" required:"true"`
	Topic   string   `envconfig:"KAFKA_TOPIC" default:"user_views"`
}

func New() (Config, error) {
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		return Config{}, fmt.Errorf("process envconfig: %w", err)
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
