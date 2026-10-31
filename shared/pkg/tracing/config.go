package tracing

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Disabled    bool    `envconfig:"OTEL_SDK_DISABLED" default:"false"`
	Endpoint    string  `envconfig:"OTEL_EXPORTER_OTLP_ENDPOINT" default:"localhost:4317"`
	Insecure    bool    `envconfig:"OTEL_EXPORTER_OTLP_INSECURE" default:"true"`
	SampleRatio float64 `envconfig:"OTEL_TRACES_SAMPLER_ARG" default:"1.0"`
}

func NewConfig() (Config, error) {
	var config Config
	if err := envconfig.Process("", &config); err != nil {
		return Config{}, fmt.Errorf("process envconfig: %w", err)
	}

	if config.Endpoint == "" {
		return Config{}, fmt.Errorf("OTEL_EXPORTER_OTLP_ENDPOINT must not be empty")
	}

	if config.SampleRatio < 0 || config.SampleRatio > 1 {
		return Config{}, fmt.Errorf(
			"OTEL_TRACES_SAMPLER_ARG must be between 0 and 1, got %f",
			config.SampleRatio,
		)
	}

	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Errorf("get tracing config: %w", err))
	}

	return config
}
