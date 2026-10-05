package redis

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
)

func Connect(ctx context.Context, config Config) (*goredis.Client, error) {
	client := goredis.NewClient(&goredis.Options{
		Addr: fmt.Sprintf("%s:%d", config.Host, config.Port),
	})

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()

		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return client, nil
}
