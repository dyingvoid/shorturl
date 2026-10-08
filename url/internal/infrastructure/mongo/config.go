package mongo

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	URI      string        `envconfig:"MONGO_URI" default:"mongodb://localhost:27017"`
	Database string        `envconfig:"MONGO_DB" required:"true"`
	Timeout  time.Duration `envconfig:"MONGO_TIMEOUT" default:"10s"`
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
