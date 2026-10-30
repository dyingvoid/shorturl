package core

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

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
	if log == nil {
		log = slog.Default()
	}

	return &AnalyticsForwarder[T]{
		log:     log,
		fetcher: fetcher,
		writer:  writer,
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
	batch, err := f.fetcher.Fetch(ctx)
	switch {
	case err == nil:
	case errors.Is(err, core_errors.ErrParseRecord):
		return err
	case errors.Is(err, context.DeadlineExceeded), errors.Is(err, context.Canceled):
		// A fetch window closing is expected: process whatever was collected so far.
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
