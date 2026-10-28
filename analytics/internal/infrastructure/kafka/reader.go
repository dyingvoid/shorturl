package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/dyingvoid/shorturl/analytics/internal/core"
	core_errors "github.com/dyingvoid/shorturl/analytics/internal/core/errors"
	"github.com/dyingvoid/shorturl/analytics/internal/domain"
)

const (
	clientID      = "analytics"
	retryBackoff  = time.Second
	initialOffset = sarama.OffsetOldest
)

var saramaVersion = sarama.V2_1_0_0

var _ core.Fetcher[domain.ClickEvent] = (*Reader)(nil)

var (
	errReaderClosed  = errors.New("kafka reader is closed")
	errSessionClosed = errors.New("kafka consumer session is closed")
)

type ackRequest struct {
	offset int64
	result chan error
}

// Reader implements core.Fetcher on top of a Sarama consumer group. Records
// are handed to the caller together with an Ack function that commits the
// corresponding partition offset.
type Reader struct {
	group   sarama.ConsumerGroup
	topic   string
	handler *consumerGroupHandler

	messages chan core.Delivery[domain.ClickEvent]
	errors   chan error

	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}

	closeOnce sync.Once
	closeErr  error

	fatalMu  sync.Mutex
	fatalErr error
}

type consumerGroupHandler struct {
	log      *slog.Logger
	messages chan<- core.Delivery[domain.ClickEvent]
	fail     func(error)
}

func NewReader(cfg Config, log *slog.Logger) (*Reader, error) {
	if log == nil {
		log = slog.Default()
	}

	saramaConfig := sarama.NewConfig()
	saramaConfig.ClientID = clientID
	saramaConfig.Version = saramaVersion
	saramaConfig.Consumer.Return.Errors = false
	saramaConfig.Consumer.Offsets.Initial = initialOffset
	saramaConfig.Consumer.Offsets.AutoCommit.Enable = false

	group, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("create consumer group: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	reader := &Reader{
		group:    group,
		topic:    cfg.Topic,
		messages: make(chan core.Delivery[domain.ClickEvent]),
		errors:   make(chan error, 1),
		ctx:      ctx,
		cancel:   cancel,
		done:     make(chan struct{}),
	}
	reader.handler = &consumerGroupHandler{
		log:      log,
		messages: reader.messages,
		fail:     reader.fail,
	}

	go reader.consume()

	return reader, nil
}

// Fetch reads up to limit deliveries. If ctx is done before the limit is
// reached, the deliveries collected so far are returned together with the
// context error (context.DeadlineExceeded for a fetch window expiring), which
// callers should treat as expected.
func (r *Reader) Fetch(ctx context.Context, limit int) ([]core.Delivery[domain.ClickEvent], error) {
	if limit <= 0 {
		return nil, fmt.Errorf("fetch: limit must be positive, got %d", limit)
	}

	deliveries := make([]core.Delivery[domain.ClickEvent], 0, limit)
	for len(deliveries) < limit {
		select {
		case delivery := <-r.messages:
			deliveries = append(deliveries, delivery)
		case err := <-r.errors:
			return deliveries, err
		case <-r.done:
			if err := r.fatal(); err != nil {
				return deliveries, err
			}
			select {
			case err := <-r.errors:
				return deliveries, err
			default:
				return deliveries, errReaderClosed
			}
		case <-ctx.Done():
			return deliveries, ctx.Err()
		}
	}

	return deliveries, nil
}

func (r *Reader) Close() error {
	r.closeOnce.Do(func() {
		r.cancel()
		<-r.done

		if err := r.group.Close(); err != nil {
			r.closeErr = fmt.Errorf("close kafka reader: %w", err)
		}
	})

	return r.closeErr
}

func (r *Reader) consume() {
	defer close(r.done)

	for {
		err := r.group.Consume(r.ctx, []string{r.topic}, r.handler)

		if r.ctx.Err() != nil || errors.Is(err, sarama.ErrClosedConsumerGroup) {
			return
		}

		if r.fatal() != nil {
			return
		}

		if err != nil {
			r.reportError(fmt.Errorf("consume kafka messages: %w", err))
		}

		select {
		case <-r.ctx.Done():
			return
		case <-time.After(retryBackoff):
		}
	}
}

func (r *Reader) fail(err error) {
	r.fatalMu.Lock()
	if r.fatalErr == nil {
		r.fatalErr = err
	}
	r.fatalMu.Unlock()

	r.reportError(err)
}

func (r *Reader) fatal() error {
	r.fatalMu.Lock()
	defer r.fatalMu.Unlock()

	return r.fatalErr
}

func (r *Reader) reportError(err error) {
	select {
	case r.errors <- err:
	default:
	}
}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) ConsumeClaim(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim,
) error {
	acks := make(chan ackRequest)

	for {
		select {
		case <-session.Context().Done():
			return nil
		case request := <-acks:
			session.MarkOffset(claim.Topic(), claim.Partition(), request.offset+1, "")
			session.Commit()
			request.result <- nil
		case message, ok := <-claim.Messages():
			if !ok {
				return nil
			}

			var event domain.ClickEvent
			if err := json.Unmarshal(message.Value, &event); err != nil {
				h.log.Error(
					"parse click event",
					"topic", message.Topic,
					"partition", message.Partition,
					"offset", message.Offset,
					"error", err,
				)
				h.fail(fmt.Errorf(
					"%w: topic=%s partition=%d offset=%d: %w",
					core_errors.ErrParseRecord,
					message.Topic,
					message.Partition,
					message.Offset,
					err,
				))
				return nil
			}

			delivery := core.Delivery[domain.ClickEvent]{
				Value: event,
				Ack: func(ctx context.Context) error {
					return commit(ctx, session.Context(), acks, message.Offset)
				},
			}

			select {
			case h.messages <- delivery:
			case <-session.Context().Done():
				return nil
			}
		}
	}
}

func commit(
	ctx context.Context,
	sessionCtx context.Context,
	acks chan<- ackRequest,
	offset int64,
) error {
	request := ackRequest{
		offset: offset,
		result: make(chan error, 1),
	}

	select {
	case acks <- request:
	case <-sessionCtx.Done():
		return errSessionClosed
	case <-ctx.Done():
		return ctx.Err()
	}

	select {
	case err := <-request.result:
		return err
	case <-sessionCtx.Done():
		return errSessionClosed
	case <-ctx.Done():
		return ctx.Err()
	}
}
