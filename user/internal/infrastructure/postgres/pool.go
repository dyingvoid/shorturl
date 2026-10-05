package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dyingvoid/shorturl/user/internal/infrastructure/postgres/errors"
)

type Pool interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	OpTimeout() time.Duration
	Close()
}

type pool struct {
	*pgxpool.Pool
	opTimeout time.Duration
}

func (p *pool) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	ctx, cancel := context.WithTimeout(ctx, p.opTimeout)
	defer cancel()

	rows, err := p.Pool.Query(ctx, sql, args...)

	return rows, mapQueryError(err)
}

func (p *pool) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	ctx, cancel := context.WithTimeout(ctx, p.opTimeout)
	defer cancel()

	return queryRow{Row: p.Pool.QueryRow(ctx, sql, args...)}
}

func (p *pool) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	ctx, cancel := context.WithTimeout(ctx, p.opTimeout)
	defer cancel()

	tag, err := p.Pool.Exec(ctx, sql, arguments...)

	return tag, mapQueryError(err)
}

func (p *pool) OpTimeout() time.Duration {
	return p.opTimeout
}

func (p *pool) Close() {
	p.Pool.Close()
}

type queryRow struct {
	pgx.Row
}

func (r queryRow) Scan(dest ...any) error {
	return mapQueryError(r.Row.Scan(dest...))
}

func mapQueryError(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == postgres_errors.PgForeignKeyViolation {
		return fmt.Errorf(
			"%v: %w",
			err,
			postgres_errors.ErrFKViolation,
		)
	}

	return err
}
