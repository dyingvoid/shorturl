package clickhouse

import (
	"context"
	"errors"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

func Connect(ctx context.Context, cfg Config) (driver.Conn, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{cfg.Addr},
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.Username,
			Password: cfg.Password,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("open clickhouse: %w", err)
	}

	if err := conn.Ping(ctx); err != nil {
		pingErr := fmt.Errorf("ping clickhouse: %w", err)
		if closeErr := conn.Close(); closeErr != nil {
			return nil, errors.Join(pingErr, fmt.Errorf("close clickhouse: %w", closeErr))
		}
		return nil, pingErr
	}

	return conn, nil
}
