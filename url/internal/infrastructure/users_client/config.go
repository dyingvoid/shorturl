package users_client

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Host           string        `envconfig:"USERS_SERVICE_HOST" default:"localhost"`
	Port           int           `envconfig:"USERS_SERVICE_PORT" required:"true"`
	RequestTimeout time.Duration `envconfig:"USERS_SERVICE_TIMEOUT" default:"5s"`
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
