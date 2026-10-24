package postgres

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Timeout  time.Duration `envconfig:"TIMEOUT" required:"true"`
	Host     string        `envconfig:"POSTGRES_HOST" default:"localhost"`
	Port     int           `envconfig:"POSTGRES_PORT" default:"5432"`
	DB       string        `envconfig:"POSTGRES_DB" required:"true"`
	User     string        `envconfig:"POSTGRES_USER" required:"true"`
	Password string        `envconfig:"POSTGRES_PASSWORD" required:"true"`
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
