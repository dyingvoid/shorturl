package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, config Config) (Pool, error) {
	pgxPool, err := pgxpool.New(ctx, dsn(config))
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	if err := pgxPool.Ping(ctx); err != nil {
		pgxPool.Close()

		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &pool{Pool: pgxPool, opTimeout: config.Timeout}, nil
}

func dsn(config Config) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.DB,
	)
}
