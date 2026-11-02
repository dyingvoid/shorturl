package redis_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/alicebob/miniredis/v2"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	domain_errors "github.com/dyingvoid/urlshort/url/internal/domain/errors"
	"github.com/dyingvoid/urlshort/url/internal/infrastructure/redis"
)

func TestSetIncCounterAtomic_DoesNotOverrideExistingCounter(t *testing.T) {
	server := miniredis.RunT(t)

	client := goredis.NewClient(&goredis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	cache := redis.NewCache(client, slog.New(slog.DiscardHandler))
	ctx := context.Background()
	userID := "user-1"
	repositoryCount := 3

	for range 2 {
		_, err := cache.GetLinkCounter(ctx, userID)
		require.ErrorIs(t, err, domain_errors.ErrNotFound)
	}

	first, err := cache.SetIncCounterAtomic(ctx, userID, repositoryCount)
	require.NoError(t, err)
	require.Equal(t, repositoryCount+1, first)

	second, err := cache.SetIncCounterAtomic(ctx, userID, repositoryCount)
	require.NoError(t, err)
	require.Equal(t, repositoryCount+2, second)

	stored, err := cache.GetLinkCounter(ctx, userID)
	require.NoError(t, err)
	require.Equal(t, repositoryCount+2, stored)
}
