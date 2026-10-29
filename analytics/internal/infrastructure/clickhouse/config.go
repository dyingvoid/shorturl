package clickhouse

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Addr     string `envconfig:"CLICKHOUSE_ADDR" default:"localhost:9000"`
	Database string `envconfig:"CLICKHOUSE_DATABASE" default:"analytics"`
	Username string `envconfig:"CLICKHOUSE_USERNAME" default:"default"`
	Password string `envconfig:"CLICKHOUSE_PASSWORD"`
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
