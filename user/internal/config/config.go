package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// HMAC-SHA256
type Config struct {
	JWTSecret       string        `envconfig:"JWT_SECRET" required:"true"`
	AccessTokenTTL  time.Duration `envconfig:"ACCESS_TOKEN_TTL" default:"15m"`
	RefreshTokenTTL time.Duration `envconfig:"REFRESH_TOKEN_TTL" default:"168h"`
	UserServicePort int           `envconfig:"USER_SERVICE_PORT" default:"50051"`
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
