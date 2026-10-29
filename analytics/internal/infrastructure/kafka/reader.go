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
) (core.MsgBatch[domain.ClickEvent], error) {
	if limit <= 0 {
		return core.MsgBatch[domain.ClickEvent]{},
			fmt.Errorf("fetch: limit must be positive, got %d", limit)
	}

	r.fetchMu.Lock()
	defer r.fetchMu.Unlock()

	if err := r.fatal(); err != nil {
		return core.MsgBatch[domain.ClickEvent]{}, err
	}

	events := make([]domain.ClickEvent, 0, limit)
	messages := make([]kafka.Message, 0, limit)

	makeBatch := func() core.MsgBatch[domain.ClickEvent] {
		if len(events) == 0 {
			return core.MsgBatch[domain.ClickEvent]{}
		}

		return core.MsgBatch[domain.ClickEvent]{
			Messages: events,
			Ack: func(ackCtx context.Context) error {
				return r.reader.CommitMessages(ackCtx, messages...)
			},
		}
	}

	for len(events) < limit {
		message, err := r.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return makeBatch(), nil
			}

			return makeBatch(), err
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

			return makeBatch(), parseErr
		}

		events = append(events, event)
		messages = append(messages, message)
	}

	return makeBatch(), nil
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
