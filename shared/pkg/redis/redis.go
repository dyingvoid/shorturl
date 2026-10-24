package redis

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type Config interface {
	Addr() string
	Timeout() time.Duration
}

func New(ctx context.Context, cfg Config) (*goredis.Client, error) {
	client := goredis.NewClient(&goredis.Options{
		Addr:         cfg.Addr(),
		DialTimeout:  cfg.Timeout(),
		ReadTimeout:  cfg.Timeout(),
		WriteTimeout: cfg.Timeout(),
	})

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()

		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return client, nil
}
