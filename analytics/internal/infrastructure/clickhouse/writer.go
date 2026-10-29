package clickhouse

import (
	"context"
	"errors"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/dyingvoid/shorturl/analytics/internal/domain"
)

const insertEventsQuery = `
INSERT INTO click_events
    (short_code, user_id, clicked_at, ip, user_agent, referer)
VALUES (?, ?, ?, ?, ?, ?)
`

type Writer struct {
	conn driver.Conn
}

func NewWriter(ctx context.Context, cfg Config) (*Writer, error) {
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

	return &Writer{conn: conn}, nil
}

func (w *Writer) WriteBatch(ctx context.Context, events []domain.ClickEvent) error {
	if len(events) == 0 {
		return nil
	}

	batch, err := w.conn.PrepareBatch(ctx, insertEventsQuery)
	if err != nil {
		return fmt.Errorf("prepare batch: %w", err)
	}

	for _, event := range events {
		if err := batch.Append(
			event.ShortCode,
			event.UserID,
			event.ClickedAt,
			event.IP,
			event.UserAgent,
			event.Referer,
		); err != nil {
			appendErr := fmt.Errorf("append click event: %w", err)
			if abortErr := batch.Abort(); abortErr != nil {
				return errors.Join(appendErr, fmt.Errorf("abort batch: %w", abortErr))
			}
			return appendErr
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("send batch: %w", err)
	}

	return nil
}

func (w *Writer) Close() error {
	if err := w.conn.Close(); err != nil {
		return fmt.Errorf("close clickhouse: %w", err)
	}

	return nil
}
