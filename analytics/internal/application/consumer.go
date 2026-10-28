package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/dyingvoid/shorturl/analytics/internal/domain"
)

const shutdownTimeout = 10 * time.Second

type Consumer struct {
	fetcher       RecordFetcher
	writer        BatchWriter
	logger        *slog.Logger
	batchSize     int
	flushInterval time.Duration

	records      []Record
	events       []domain.ClickEvent
	batchStarted time.Time
}

func NewConsumer(
	fetcher RecordFetcher,
	writer BatchWriter,
	batchSize int,
	flushInterval time.Duration,
	logger *slog.Logger,
) *Consumer {
	if logger == nil {
		logger = slog.Default()
	}

	return &Consumer{
		fetcher:       fetcher,
		writer:        writer,
		logger:        logger,
		batchSize:     batchSize,
		flushInterval: flushInterval,
	}
}

func (c *Consumer) Run(ctx context.Context) error {
	for {
		if ctx.Err() != nil {
			return c.shutdown()
		}

		if c.eventLimitReached() {
			if err := c.flush(ctx); err != nil {
				return err
			}
		}

		if c.flushDeadlineReached() {
			if err := c.flush(ctx); err != nil {
				return err
			}
			continue
		}

		fetchCtx, cancel := c.fetchContext(ctx)
		record, err := c.fetcher.FetchMessage(fetchCtx)
		timedOut := errors.Is(fetchCtx.Err(), context.DeadlineExceeded)
		cancel()

		if err != nil {
			if ctx.Err() != nil {
				return c.shutdown()
			}
			if timedOut || errors.Is(err, context.DeadlineExceeded) {
				if len(c.records) > 0 {
					if err := c.flush(ctx); err != nil {
						return err
					}
				}
				continue
			}
			return fmt.Errorf("fetch kafka message: %w", err)
		}

		c.add(record)

		if c.eventLimitReached() {
			if err := c.flush(ctx); err != nil {
				return err
			}
		}
	}
}

func (c *Consumer) add(record Record) {
	if len(c.records) == 0 {
		c.batchStarted = time.Now()
	}

	c.records = append(c.records, record)

	var event domain.ClickEvent
	if err := json.Unmarshal(record.Value, &event); err != nil {
		c.logger.Warn(
			"skipping malformed click event",
			"topic", record.Topic,
			"partition", record.Partition,
			"offset", record.Offset,
			"error", err,
		)
		return
	}

	c.events = append(c.events, event)
}

func (c *Consumer) flush(ctx context.Context) error {
	if len(c.records) == 0 {
		return nil
	}

	records := c.records
	events := c.events

	if len(events) == 0 {
		c.logger.Info(
			"batch has no valid events, committing skipped records",
			"records", len(records),
		)
	} else if err := c.writer.WriteBatch(ctx, events); err != nil {
		return fmt.Errorf("write click events: %w", err)
	}

	if err := c.fetcher.CommitMessages(ctx, records...); err != nil {
		return fmt.Errorf("commit kafka offsets: %w", err)
	}

	c.records = nil
	c.events = nil
	c.batchStarted = time.Time{}

	c.logger.Debug(
		"batch persisted and offsets committed",
		"records", len(records),
		"events", len(events),
	)

	return nil
}

func (c *Consumer) shutdown() error {
	if len(c.records) == 0 {
		c.logger.Info("consumer stopped")
		return nil
	}

	c.logger.Info(
		"flushing remaining batch during shutdown",
		"records", len(c.records),
		"events", len(c.events),
	)

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := c.flush(ctx); err != nil {
		return fmt.Errorf("flush batch during shutdown: %w", err)
	}

	c.logger.Info("consumer stopped")
	return nil
}

func (c *Consumer) eventLimitReached() bool {
	return len(c.events) >= c.batchSize
}

func (c *Consumer) flushDeadlineReached() bool {
	return !c.batchStarted.IsZero() && time.Since(c.batchStarted) >= c.flushInterval
}

func (c *Consumer) fetchContext(ctx context.Context) (context.Context, context.CancelFunc) {
	timeout := c.flushInterval
	if !c.batchStarted.IsZero() {
		timeout = c.flushInterval - time.Since(c.batchStarted)
		if timeout < 0 {
			timeout = 0
		}
	}

	return context.WithTimeout(ctx, timeout)
}
