package config

import (
	"fmt"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	TimeZone              *time.Location
	CreateURLAttempts     int           `envconfig:"URL_CREATE_ATTEMPTS" default:"3"`
	ServerPort            int           `envconfig:"URL_SERVICE_PORT" default:"50051"`
	ServerShutdownTimeout time.Duration `envconfig:"GRPC_SHUTDOWN_TIMEOUT" default:"10s"`
}

func New() (Config, error) {
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		return Config{}, fmt.Errorf("process envconfig: %w", err)
	}

	tz := os.Getenv("TIME_ZONE")
	if tz == "" {
		tz = "UTC"
	}

	zone, err := time.LoadLocation(tz)
	if err != nil {
		return Config{}, fmt.Errorf("load time zone: %s: %w", tz, err)
	}
	config.TimeZone = zone

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

func (c Config) Port() int {
	return c.ServerPort
}

func (c Config) ShutdownTimeout() time.Duration {
	return c.ServerShutdownTimeout
}
