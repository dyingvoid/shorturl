package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"sync"
	"time"

	kafka "github.com/segmentio/kafka-go"

	"github.com/dyingvoid/shorturl/analytics/internal/core"
	core_errors "github.com/dyingvoid/shorturl/analytics/internal/core/errors"
	"github.com/dyingvoid/shorturl/analytics/internal/domain"
)

var clickEventBytes = int(reflect.TypeOf(domain.ClickEvent{}).Size())

const (
	defaultBatchSize  = 100
	defaultBatchFlush = 5 * time.Second
)

type Reader struct {
	cfg    Config
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
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = defaultBatchSize
	}
	if cfg.BatchFlushInterval <= 0 {
		cfg.BatchFlushInterval = defaultBatchFlush
	}

	readerConfig := kafka.ReaderConfig{
		Brokers:     cfg.Brokers,
		GroupID:     cfg.GroupID,
		Topic:       cfg.Topic,
		StartOffset: kafka.FirstOffset,

		MinBytes:      1,
		MaxBytes:      clickEventBytes * cfg.BatchSize,
		MaxWait:       cfg.BatchFlushInterval,
		QueueCapacity: 128,

		CommitInterval: 0,
	}

	if err := readerConfig.Validate(); err != nil {
		return nil, fmt.Errorf("validate kafka config: %w", err)
	}

	return &Reader{
		cfg:    cfg,
		reader: kafka.NewReader(readerConfig),
		log:    log,
	}, nil
}

func (r *Reader) Fetch(ctx context.Context) (core.MsgBatch[domain.ClickEvent], error) {
	limit := r.cfg.BatchSize
	if limit <= 0 {
		limit = 1
	}

	fetchCtx, cancel := context.WithTimeout(ctx, r.cfg.BatchFlushInterval)
	defer cancel()

	r.fetchMu.Lock()
	defer r.fetchMu.Unlock()

	if err := r.fatal(); err != nil {
		return core.MsgBatch[domain.ClickEvent]{}, err
	}

	events := make([]domain.ClickEvent, 0, limit)
	messages := make([]kafka.Message, 0, limit)

	for len(events) < limit {
		message, err := r.reader.FetchMessage(fetchCtx)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return r.newBatch(events, messages), nil
			}

			return r.newBatch(events, messages), err
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

			return r.newBatch(events, messages), parseErr
		}

		events = append(events, event)
		messages = append(messages, message)
	}

	return r.newBatch(events, messages), nil
}

func (r *Reader) newBatch(events []domain.ClickEvent, messages []kafka.Message) core.MsgBatch[domain.ClickEvent] {
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
