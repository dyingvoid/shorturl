package redis

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	domain_errors "github.com/dyingvoid/shorturl/user/internal/domain/errors"
	goredis "github.com/redis/go-redis/v9"
)

type Cache struct {
	client *goredis.Client
	logger *slog.Logger
}

func NewCache(client *goredis.Client, logger *slog.Logger) *Cache {
	return &Cache{client: client, logger: logger}
}

func (c *Cache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	if err := c.client.Set(ctx, key, value, ttl).Err(); err != nil {
		wrapped := fmt.Errorf("set cache key '%s': %w", key, err)
		c.logger.Error("set cache key", "error", wrapped, "key", key)

		return wrapped
	}

	return nil
}

func (c *Cache) Get(ctx context.Context, key string) ([]byte, error) {
	value, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			c.logger.Info("cache miss", "key", key)

			return nil, fmt.Errorf("get cache key '%s': %w", key, domain_errors.ErrCacheMiss)
		}

		wrapped := fmt.Errorf("get cache key '%s': %w", key, err)
		c.logger.Error("get cache key", "error", wrapped, "key", key)

		return nil, wrapped
	}

	return value, nil
}

func (c *Cache) Delete(ctx context.Context, key string) error {
	if err := c.client.Del(ctx, key).Err(); err != nil {
		wrapped := fmt.Errorf("delete cache key '%s': %w", key, err)
		c.logger.Error("delete cache key", "error", wrapped, "key", key)

		return wrapped
	}

	return nil
}
