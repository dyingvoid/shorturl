package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	UsersServiceHost string `envconfig:"USER_SERVICE_HOST" default:"localhost"`
	UsersServicePort int    `envconfig:"USER_SERVICE_PORT" default:"50051"`
	URLServiceHost   string `envconfig:"URL_SERVICE_HOST" default:"localhost"`
	URLServicePort   int    `envconfig:"URL_SERVICE_PORT" default:"50052"`
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
		panic(fmt.Errorf("get config: %w", err))
	}

	return config
}

func (c Config) UsersServiceAddr() string {
	return fmt.Sprintf("%s:%d", c.UsersServiceHost, c.UsersServicePort)
}

func (c Config) URLServiceAddr() string {
	return fmt.Sprintf("%s:%d", c.URLServiceHost, c.URLServicePort)
}
