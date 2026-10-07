package postgres

import (
	"context"
	"time"

	postgres_errors "github.com/dyingvoid/shorturl/user/internal/infrastructure/postgres/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Pool interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Begin(ctx context.Context) (Tx, error)
	OpTimeout() time.Duration
	Close()
}

type Tx interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type pool struct {
	*pgxpool.Pool
	opTimeout time.Duration
}

func (p *pool) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	tag, err := p.Pool.Exec(ctx, sql, arguments...)

	return tag, postgres_errors.MapError(err)
}

func (p *pool) Begin(ctx context.Context) (Tx, error) {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}

	return &txWrapper{Tx: tx}, nil
}

type txWrapper struct {
	pgx.Tx
}

func (t *txWrapper) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	tag, err := t.Tx.Exec(ctx, sql, arguments...)

	return tag, postgres_errors.MapError(err)
}

func (p *pool) OpTimeout() time.Duration {
	return p.opTimeout
}

func (p *pool) Close() {
	p.Pool.Close()
}
