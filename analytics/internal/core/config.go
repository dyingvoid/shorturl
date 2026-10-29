package core

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	BatchSize          int           `envconfig:"BATCH_SIZE" default:"100"`
	BatchFlushInterval time.Duration `envconfig:"BATCH_FLUSH_INTERVAL" default:"5s"`
	DefaultDays        int32         `envconfig:"ANALYTICS_DEFAULT_DAYS" default:"30"`
	MaxDays            int32         `envconfig:"ANALYTICS_MAX_DAYS" default:"365"`
	TopReferers        int           `envconfig:"ANALYTICS_TOP_REFERERS" default:"10"`
}

func New() (Config, error) {
	var config Config
	if err := envconfig.Process("", &config); err != nil {
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
