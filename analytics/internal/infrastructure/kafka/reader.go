package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	kafka "github.com/segmentio/kafka-go"

	"github.com/dyingvoid/shorturl/analytics/internal/core"
	core_errors "github.com/dyingvoid/shorturl/analytics/internal/core/errors"
	"github.com/dyingvoid/shorturl/analytics/internal/domain"
)

const (
	maxBytes = 10e6
	maxWait  = 250 * time.Millisecond
)

type Reader struct {
	reader *kafka.Reader
	log    *slog.Logger

	fetchMu sync.Mutex

	fatalMu  sync.RWMutex
	fatalErr error

	closeOnce sync.Once
	closeErr  error
}

func NewReader(cfg Config, log *slog.Logger) (*Reader, error) {
	if log == nil {
		log = slog.Default()
	}

	readerConfig := kafka.ReaderConfig{
		Brokers:     cfg.Brokers,
		GroupID:     cfg.GroupID,
		Topic:       cfg.Topic,
		StartOffset: kafka.FirstOffset,

		MinBytes:      1,
		MaxBytes:      maxBytes,
		MaxWait:       maxWait,
		QueueCapacity: 128,

		CommitInterval: 0,
	}

	if err := readerConfig.Validate(); err != nil {
		return nil, fmt.Errorf("validate kafka config: %w", err)
	}

	return &Reader{
		reader: kafka.NewReader(readerConfig),
		log:    log,
	}, nil
}

func (r *Reader) Fetch(
	ctx context.Context,
	limit int,
) ([]core.Delivery[domain.ClickEvent], error) {
	if limit <= 0 {
		return nil, fmt.Errorf("fetch: limit must be positive, got %d", limit)
	}

	r.fetchMu.Lock()
	defer r.fetchMu.Unlock()

	if err := r.fatal(); err != nil {
		return nil, err
	}

	deliveries := make(
		[]core.Delivery[domain.ClickEvent],
		0,
		limit,
	)

	for len(deliveries) < limit {
		message, err := r.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return deliveries, nil
			}

			return deliveries, err
		}

		var event domain.ClickEvent
		if err := json.Unmarshal(message.Value, &event); err != nil {
			parseErr := fmt.Errorf(
				"%w: topic=%s partition=%d offset=%d: %v",
				core_errors.ErrParseRecord,
				message.Topic,
				message.Partition,
				message.Offset,
				err,
			)

			r.log.Error("parse click event", "error", parseErr)
			r.setFatal(parseErr)

			return deliveries, parseErr
		}

		msg := message

		deliveries = append(deliveries, core.Delivery[domain.ClickEvent]{
			Value: event,
			Ack: func(ackCtx context.Context) error {
				return r.reader.CommitMessages(ackCtx, msg)
			},
		})
	}

	return deliveries, nil
}

func (r *Reader) fatal() error {
	r.fatalMu.RLock()
	defer r.fatalMu.RUnlock()

	return r.fatalErr
}

func (r *Reader) setFatal(err error) {
	r.fatalMu.Lock()
	defer r.fatalMu.Unlock()

	if r.fatalErr == nil {
		r.fatalErr = err
	}
}

func (r *Reader) Close() error {
	r.closeOnce.Do(func() {
		if err := r.reader.Close(); err != nil {
			r.closeErr = fmt.Errorf("close kafka reader: %w", err)
		}
	})

	return r.closeErr
}