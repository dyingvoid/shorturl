package clickhouse

import (
	"errors"
	"strings"
)

type Config struct {
	Addr     string
	Database string
	Username string
	Password string
}

func NewConfig(addr, database, username, password string) (Config, error) {
	cfg := Config{
		Addr:     strings.TrimSpace(addr),
		Database: strings.TrimSpace(database),
		Username: strings.TrimSpace(username),
		Password: password,
	}

	if cfg.Addr == "" {
		return Config{}, errors.New("clickhouse: address is required")
	}
	if cfg.Database == "" {
		return Config{}, errors.New("clickhouse: database is required")
	}
	if cfg.Username == "" {
		return Config{}, errors.New("clickhouse: username is required")
	}

	return cfg, nil
}
