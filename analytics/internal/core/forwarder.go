package core

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	core_errors "github.com/dyingvoid/shorturl/analytics/internal/core/errors"
)

type AnalyticsForwarder[T any] struct {
	log     *slog.Logger
	fetcher Fetcher[T]
	writer  Writer[T]
	metrics Metrics
}

func NewAnalyticsForwarder[T any](
	log *slog.Logger,
	fetcher Fetcher[T],
	writer Writer[T],
	metrics Metrics,
) *AnalyticsForwarder[T] {
	if log == nil {
		log = slog.Default()
	}

	return &AnalyticsForwarder[T]{
		log:     log,
		fetcher: fetcher,
		writer:  writer,
		metrics: metrics,
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
	default:
		f.log.Error("fetch records", "error", err)
	}

	if len(batch.Messages) == 0 {
		return nil
	}

	start := time.Now()
	writeErr := f.writer.WriteBatch(ctx, batch.Messages)
	f.metrics.ObserveClickHouseBatchInsertDuration(time.Since(start).Seconds())

	if writeErr != nil {
		return fmt.Errorf("write batch: %w", writeErr)
	}

	f.metrics.IncKafkaMessagesConsumed(len(batch.Messages))
	batch.Ack(ctx)
	f.log.Debug("batch processed", "records", len(batch.Messages))
	return nil
}
