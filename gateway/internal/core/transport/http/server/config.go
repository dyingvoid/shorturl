package core_http_server

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Addr              string        `ignored:"true"`
	Port              int           `envconfig:"PORT" default:"8082"`
	MetricsPort       int           `envconfig:"METRICS_PORT" default:"9090"`
	ShutdownTimeout   time.Duration `envconfig:"SHUTDOWN_TIMEOUT" default:"10s"`
	RequestTimeout    time.Duration `envconfig:"REQUEST_TIMEOUT" default:"5s"`
	AllowedOrigins    []string      `envconfig:"ALLOWED_ORIGINS" default:"http://localhost:3000"`
	ReadHeaderTimeout time.Duration `envconfig:"READ_HEADER_TIMEOUT" default:"5s"`
	ReadTimeout       time.Duration `envconfig:"READ_TIMEOUT" default:"10s"`
	WriteTimeout      time.Duration `envconfig:"WRITE_TIMEOUT" default:"15s"`
	IdleTimeout       time.Duration `envconfig:"IDLE_TIMEOUT" default:"60s"`
}

func NewConfig() (Config, error) {
	var config Config
	if err := envconfig.Process("GATEWAY", &config); err != nil {
		return Config{}, fmt.Errorf("process envconfig: %w", err)
	}

	config.Addr = fmt.Sprintf(":%d", config.Port)
	return config, nil
}

func MustNewConfig() Config {
	config, err := NewConfig()
	if err != nil {
		panic(fmt.Errorf("get config: %w", err))
	}

	return config
}

func (c Config) Address() string {
	if c.Addr != "" {
		return c.Addr
	}

	return fmt.Sprintf(":%d", c.Port)
}
