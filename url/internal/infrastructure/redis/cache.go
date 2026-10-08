package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/dyingvoid/urlshort/url/internal/domain"
	domain_errors "github.com/dyingvoid/urlshort/url/internal/domain/errors"
	goredis "github.com/redis/go-redis/v9"
)

const (
	linkCounterKeyPrefix = "counter:"
	linkKeyPrefix        = "link:"
)

type Cache struct {
	client *goredis.Client
	logger *slog.Logger
}

func NewCache(client *goredis.Client, logger *slog.Logger) *Cache {
	return &Cache{client: client, logger: logger}
}

func (c *Cache) GetLinkCounter(ctx context.Context, userID string) (int, error) {
	key := linkCounterKey(userID)

	value, err := c.client.Get(ctx, key).Int()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			c.logger.Info("cache miss", "key", key)

			return 0, fmt.Errorf("get cache key '%s': %w", key, domain_errors.ErrNotFound)
		}

		wrapped := fmt.Errorf("get cache key '%s': %w", key, err)
		c.logger.Error("get cache key", "error", wrapped, "key", key)

		return 0, wrapped
	}

	return value, nil
}

func (c *Cache) SetLinkCounter(ctx context.Context, userID string, val int) error {
	key := linkCounterKey(userID)

	if err := c.client.Set(ctx, key, val, 0).Err(); err != nil {
		wrapped := fmt.Errorf("set cache key '%s': %w", key, err)
		c.logger.Error("set cache key", "error", wrapped, "key", key)

		return wrapped
	}

	return nil
}

func (c *Cache) IncLinkCounter(ctx context.Context, userID string) (int, error) {
	key := linkCounterKey(userID)

	value, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		wrapped := fmt.Errorf("increment cache key '%s': %w", key, err)
		c.logger.Error("increment cache key", "error", wrapped, "key", key)

		return 0, wrapped
	}

	return int(value), nil
}

func (c *Cache) DecLinkCounter(ctx context.Context, userID string) (int, error) {
	key := linkCounterKey(userID)

	value, err := c.client.Decr(ctx, key).Result()
	if err != nil {
		wrapped := fmt.Errorf("decrement cache key '%s': %w", key, err)
		c.logger.Error("decrement cache key", "error", wrapped, "key", key)

		return 0, wrapped
	}

	return int(value), nil
}

func (c *Cache) GetLink(ctx context.Context, shortCode string) (*domain.Link, error) {
	key := linkKey(shortCode)

	value, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			c.logger.Info("cache miss", "key", key)

			return nil, fmt.Errorf("get cache key '%s': %w", key, domain_errors.ErrNotFound)
		}

		wrapped := fmt.Errorf("get cache key '%s': %w", key, err)
		c.logger.Error("get cache key", "error", wrapped, "key", key)

		return nil, wrapped
	}

	var link domain.Link
	if err := json.Unmarshal(value, &link); err != nil {
		wrapped := fmt.Errorf("unmarshal cache key '%s': %w", key, err)
		c.logger.Error("unmarshal cache key", "error", wrapped, "key", key)

		return nil, wrapped
	}

	return &link, nil
}

func (c *Cache) SetLink(ctx context.Context, link *domain.Link, ttl time.Duration) error {
	key := linkKey(link.ShortCode)

	value, err := json.Marshal(link)
	if err != nil {
		wrapped := fmt.Errorf("marshal cache key '%s': %w", key, err)
		c.logger.Error("marshal cache key", "error", wrapped, "key", key)

		return wrapped
	}

	if err := c.client.Set(ctx, key, value, ttl).Err(); err != nil {
		wrapped := fmt.Errorf("set cache key '%s': %w", key, err)
		c.logger.Error("set cache key", "error", wrapped, "key", key)

		return wrapped
	}

	return nil
}

func (c *Cache) DeleteLink(ctx context.Context, shortCode string) error {
	key := linkKey(shortCode)

	if err := c.client.Del(ctx, key).Err(); err != nil {
		wrapped := fmt.Errorf("delete cache key '%s': %w", key, err)
		c.logger.Error("delete cache key", "error", wrapped, "key", key)

		return wrapped
	}

	return nil
}

func linkKey(shortCode string) string {
	return fmt.Sprintf("%s%s", linkKeyPrefix, shortCode)
}

func linkCounterKey(userID string) string {
	return fmt.Sprintf("%s%s", linkCounterKeyPrefix, userID)
}
