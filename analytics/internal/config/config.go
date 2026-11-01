package config

import (
	"fmt"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	TimeZone              *time.Location
	ServerPort            int           `envconfig:"SERVICE_PORT" default:"50053"`
	MetricsServerPort     int           `envconfig:"METRICS_PORT" default:"9090"`
	ServerShutdownTimeout time.Duration `envconfig:"GRPC_SHUTDOWN_TIMEOUT" default:"10s"`
	DefaultDays           int32         `envconfig:"DEFAULT_DAYS" default:"30"`
	MaxDays               int32         `envconfig:"MAX_DAYS" default:"365"`
	TopReferers           int           `envconfig:"TOP_REFERERS" default:"10"`
}

func New() (Config, error) {
	var config Config
	if err := envconfig.Process("ANALYTICS", &config); err != nil {
		return Config{}, fmt.Errorf("process envconfig: %w", err)
	}

	switch {
	case config.DefaultDays <= 0:
		return Config{}, fmt.Errorf("ANALYTICS_DEFAULT_DAYS must be positive, got %d", config.DefaultDays)
	case config.MaxDays < config.DefaultDays:
		return Config{}, fmt.Errorf("ANALYTICS_MAX_DAYS must be greater than or equal to ANALYTICS_DEFAULT_DAYS, got %d and %d", config.MaxDays, config.DefaultDays)
	case config.TopReferers <= 0:
		return Config{}, fmt.Errorf("ANALYTICS_TOP_REFERERS must be positive, got %d", config.TopReferers)
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

func (c Config) MetricsPort() int {
	return c.MetricsServerPort
}

func (c Config) ShutdownTimeout() time.Duration {
	return c.ServerShutdownTimeout
}
