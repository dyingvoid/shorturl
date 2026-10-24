package mongo

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Timeout    time.Duration `envconfig:"MONGO_TIMEOUT" default:"10s"`
	Host       string        `envconfig:"MONGO_HOST" default:"localhost"`
	Port       int           `envconfig:"MONGO_PORT" default:"27017"`
	DB         string        `envconfig:"MONGO_DB" required:"true"`
	User       string        `envconfig:"MONGO_USER" required:"true"`
	Password   string        `envconfig:"MONGO_PASSWORD" required:"true"`
	AuthSource string        `envconfig:"MONGO_AUTH_SOURCE" default:"admin"`
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
