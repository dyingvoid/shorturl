package clickhouse

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/dyingvoid/shorturl/analytics/internal/domain"
)

const topReferersLimit = 10

type Repository struct {
	conn driver.Conn
}

func NewRepository(conn driver.Conn) *Repository {
	return &Repository{conn: conn}
}

func (r *Repository) WriteBatch(ctx context.Context, events []domain.ClickEvent) error {
	if len(events) == 0 {
		return nil
	}

	const query = `
	INSERT INTO click_events
		(short_code, user_id, clicked_at, ip, user_agent, referer)
	VALUES (?, ?, ?, ?, ?, ?)
	`

	batch, err := r.conn.PrepareBatch(ctx, query)
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

func (r *Repository) TotalClicks(
	ctx context.Context,
	shortCode string,
) (int64, error) {
	const query = `
	SELECT count()
	FROM click_events
	WHERE short_code = ?
	`

	var total uint64
	if err := r.conn.QueryRow(ctx, query, shortCode).Scan(&total); err != nil {
		return 0, fmt.Errorf("query total clicks: %w", err)
	}

	return int64(total), nil
}

func (r *Repository) ClicksByDay(
	ctx context.Context,
	shortCode string,
	from time.Time,
) ([]domain.DayClicks, error) {
	const query = `
	SELECT toDate(clicked_at) AS d, count() AS c
	FROM click_events
	WHERE short_code = ? AND clicked_at >= ?
	GROUP BY d
	ORDER BY d DESC
	`

	rows, err := r.conn.Query(ctx, query, shortCode, from)
	if err != nil {
		return nil, fmt.Errorf("query clicks by day: %w", err)
	}
	defer rows.Close()

	result := make([]domain.DayClicks, 0)
	for rows.Next() {
		var day time.Time
		var clicks uint64

		if err := rows.Scan(&day, &clicks); err != nil {
			return nil, fmt.Errorf("scan clicks by day: %w", err)
		}

		result = append(result, domain.DayClicks{
			Date:   day,
			Clicks: int64(clicks),
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate clicks by day: %w", err)
	}

	return result, nil
}

func (r *Repository) TopReferers(
	ctx context.Context,
	shortCode string,
	from time.Time,
	limit int,
) ([]domain.RefererClicks, error) {
	if limit <= 0 {
		limit = topReferersLimit
	}

	const query = `
	SELECT referer, count() AS c
	FROM click_events
	WHERE short_code = ? AND clicked_at >= ?
	GROUP BY referer
	ORDER BY c DESC
	LIMIT ?
	`

	rows, err := r.conn.Query(ctx, query, shortCode, from, limit)
	if err != nil {
		return nil, fmt.Errorf("query top referers: %w", err)
	}
	defer rows.Close()

	result := make([]domain.RefererClicks, 0)
	for rows.Next() {
		var referer string
		var clicks uint64

		if err := rows.Scan(&referer, &clicks); err != nil {
			return nil, fmt.Errorf("scan top referers: %w", err)
		}

		result = append(result, domain.RefererClicks{
			Referer: referer,
			Clicks:  int64(clicks),
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate top referers: %w", err)
	}

	return result, nil
}
