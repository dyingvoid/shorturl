package core

import (
	"context"
	"errors"
	"log/slog"
	"time"

	core_errors "github.com/dyingvoid/shorturl/analytics/internal/core/errors"
)

type AnalyticsForwarder[T any] struct {
	log     *slog.Logger
	fetcher Fetcher[T]
	writer  Writer[T]
}

func NewAnalyticsForwarder[T any](
	log *slog.Logger,
	fetcher Fetcher[T],
	writer Writer[T],
) *AnalyticsForwarder[T] {
	return &AnalyticsForwarder[T]{
		log:     log,
		fetcher: fetcher,
		writer:  writer,
	}
}

func (f AnalyticsForwarder[T]) Run(ctx context.Context) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		if err := f.runOnce(ctx); err != nil {
			if errors.Is(err, core_errors.ErrParseRecord) {
				return err
			}
			if errors.Is(err, context.DeadlineExceeded) {
				continue
			}
			// TODO: log
		}
	}
}

func (f AnalyticsForwarder[T]) runOnce(ctx context.Context) error {
	// TODO: move limit and timeout values to config
	const limit = 100
	const timeout = 5 * time.Second

	tctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	deliveries, err := f.fetcher.Fetch(tctx, limit)
	if err != nil {
		if errors.Is(err, core_errors.ErrParseRecord) {
			return err
		}
		if errors.Is(err, context.DeadlineExceeded) {
			// expected
		} else {
			// TODO: just log and continue
		}
	}

	values := make([]T, 0, len(deliveries))
	for _, d := range deliveries {
		values = append(values, d.Value)
	}

	f.writer.WriteBatch(tctx, values)
	return nil
}
