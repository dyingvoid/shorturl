package redis

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Host           string        `envconfig:"REDIS_HOST" required:"true"`
	Port           int           `envconfig:"REDIS_PORT" default:"6379"`
	RequestTimeout time.Duration `envconfig:"REDIS_TIMEOUT" required:"true"`
}

func (c Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func (c Config) Timeout() time.Duration {
	return c.RequestTimeout
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
