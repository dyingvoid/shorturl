package core

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	core_errors "github.com/dyingvoid/shorturl/analytics/internal/core/errors"
)

const (
	defaultLimit   = 100
	defaultTimeout = 5 * time.Second
)

type AnalyticsForwarder[T any] struct {
	log     *slog.Logger
	fetcher Fetcher[T]
	writer  Writer[T]
	limit   int
	timeout time.Duration
}

func NewAnalyticsForwarder[T any](
	log *slog.Logger,
	fetcher Fetcher[T],
	writer Writer[T],
	limit int,
	timeout time.Duration,
) *AnalyticsForwarder[T] {
	if log == nil {
		log = slog.Default()
	}
	if limit <= 0 {
		limit = defaultLimit
	}
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	return &AnalyticsForwarder[T]{
		log:     log,
		fetcher: fetcher,
		writer:  writer,
		limit:   limit,
		timeout: timeout,
	}
}

func (f AnalyticsForwarder[T]) Run(ctx context.Context) error {
	for {
		if ctx.Err() != nil {
			return nil
		}

		if err := f.runOnce(ctx); err != nil {
			if errors.Is(err, core_errors.ErrParseRecord) {
				return err
			}

			f.log.Error("process batch", "error", err)
		}
	}
}

func (f AnalyticsForwarder[T]) runOnce(ctx context.Context) error {
	tctx, cancel := context.WithTimeout(ctx, f.timeout)
	defer cancel()

	batch, err := f.fetcher.Fetch(tctx, f.limit)
	switch {
	case err == nil:
	case errors.Is(err, core_errors.ErrParseRecord):
		return err
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
		// A fetch window closing without reaching the limit is expected:
		// process whatever was collected so far.
	default:
		f.log.Error("fetch records", "error", err)
	}

	if len(batch.Messages) == 0 {
		return nil
	}

	if err := f.writer.WriteBatch(ctx, batch.Messages); err != nil {
		return fmt.Errorf("write batch: %w", err)
	}

	batch.Ack(ctx)
	f.log.Debug("batch processed", "records", len(batch.Messages))
	return nil
}
