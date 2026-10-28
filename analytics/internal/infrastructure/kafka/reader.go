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

// TODO: to .env, config.go struct Config with "envconfig" struct fields
const (
	clientID          = "analytics"
	retryBackoff      = time.Second
	messageBufferSize = 128
	errorBufferSize   = 8
)

var (
	saramaVersion = sarama.V2_1_0_0

	errReaderClosed  = errors.New("kafka reader is closed")
	errSessionClosed = errors.New("kafka consumer session is closed")
)

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

	fatalMu  sync.RWMutex
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
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	saramaConfig.Consumer.Offsets.AutoCommit.Enable = false

	group, err := sarama.NewConsumerGroup(
		cfg.Brokers,
		cfg.GroupID,
		saramaConfig,
	)
	if err != nil {
		return nil, fmt.Errorf("create consumer group: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	reader := &Reader{
		group:    group,
		topic:    cfg.Topic,
		messages: make(chan core.Delivery[domain.ClickEvent], messageBufferSize),
		errors:   make(chan error, errorBufferSize),
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

func (r *Reader) Fetch(
	ctx context.Context,
	limit int,
) ([]core.Delivery[domain.ClickEvent], error) {
	if limit <= 0 {
		return nil, fmt.Errorf("fetch: limit must be positive, got %d", limit)
	}

	deliveries := make([]core.Delivery[domain.ClickEvent], 0, limit)

	for len(deliveries) < limit {
		if err := ctx.Err(); err != nil {
			return deliveries, err
		}

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
			r.closeErr = fmt.Errorf("close kafka consumer group: %w", err)
		}
	})

	return r.closeErr
}

func (r *Reader) consume() {
	defer close(r.done)

	for {
		err := r.group.Consume(
			r.ctx,
			[]string{r.topic},
			r.handler,
		)

		if r.ctx.Err() != nil || errors.Is(err, sarama.ErrClosedConsumerGroup) {
			return
		}

		if r.fatal() != nil {
			return
		}

		if err == nil {
			continue
		}

		r.reportError(fmt.Errorf("consume kafka messages: %w", err))

		if !r.waitForRetry() {
			return
		}
	}
}

func (r *Reader) waitForRetry() bool {
	timer := time.NewTimer(retryBackoff)
	defer timer.Stop()

	select {
	case <-r.ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func (r *Reader) fail(err error) {
	r.fatalMu.Lock()

	if r.fatalErr != nil {
		r.fatalMu.Unlock()
		return
	}

	r.fatalErr = err
	r.fatalMu.Unlock()

	r.reportError(err)
	r.cancel()
}

func (r *Reader) fatal() error {
	r.fatalMu.RLock()
	defer r.fatalMu.RUnlock()

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
	acks := newAckTracker(session, claim)
	defer acks.Close()

	for {
		select {
		case <-session.Context().Done():
			return nil

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

			offset := message.Offset

			if err := acks.Register(offset); err != nil {
				return nil
			}

			delivery := core.Delivery[domain.ClickEvent]{
				Value: event,
				Ack: func(ctx context.Context) error {
					return acks.Ack(ctx, offset)
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

type ackRequest struct {
	offset int64
	result chan error
}

type pendingAck struct {
	offset  int64
	acked   bool
	waiters []chan error
}

type ackTracker struct {
	session   sarama.ConsumerGroupSession
	topic     string
	partition int32

	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}

	register chan int64
	requests chan ackRequest
}

func newAckTracker(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim,
) *ackTracker {
	ctx, cancel := context.WithCancel(session.Context())

	tracker := &ackTracker{
		session:   session,
		topic:     claim.Topic(),
		partition: claim.Partition(),
		ctx:       ctx,
		cancel:    cancel,
		done:      make(chan struct{}),
		register:  make(chan int64),
		requests:  make(chan ackRequest),
	}

	go tracker.run()

	return tracker
}

func (t *ackTracker) Register(offset int64) error {
	select {
	case t.register <- offset:
		return nil
	case <-t.ctx.Done():
		return errSessionClosed
	case <-t.done:
		return errSessionClosed
	}
}

func (t *ackTracker) Ack(ctx context.Context, offset int64) error {
	request := ackRequest{
		offset: offset,
		result: make(chan error, 1),
	}

	select {
	case t.requests <- request:
	case <-ctx.Done():
		return ctx.Err()
	case <-t.ctx.Done():
		return errSessionClosed
	case <-t.done:
		return errSessionClosed
	}

	select {
	case err := <-request.result:
		return err
	case <-ctx.Done():
		return ctx.Err()
	case <-t.ctx.Done():
		return errSessionClosed
	case <-t.done:
		return errSessionClosed
	}
}

func (t *ackTracker) Close() {
	t.cancel()
	<-t.done
}

func (t *ackTracker) run() {
	queue := make([]*pendingAck, 0, 128)
	pending := make(map[int64]*pendingAck)
	head := 0

	var committedOffset int64
	hasCommitted := false

	defer func() {
		for i := head; i < len(queue); i++ {
			for _, waiter := range queue[i].waiters {
				waiter <- errSessionClosed
			}
		}

		close(t.done)
	}()

	for {
		select {
		case <-t.ctx.Done():
			return

		case offset := <-t.register:
			item := &pendingAck{offset: offset}
			pending[offset] = item
			queue = append(queue, item)

		case request := <-t.requests:
			item, ok := pending[request.offset]
			if !ok {
				if hasCommitted && request.offset <= committedOffset {
					request.result <- nil
				} else {
					request.result <- fmt.Errorf(
						"ack offset %d is not pending",
						request.offset,
					)
				}
				continue
			}

			item.acked = true
			item.waiters = append(item.waiters, request.result)

			start := head

			for head < len(queue) && queue[head].acked {
				delete(pending, queue[head].offset)
				head++
			}

			if start == head {
				continue
			}

			lastOffset := queue[head-1].offset

			t.session.MarkOffset(
				t.topic,
				t.partition,
				lastOffset+1,
				"",
			)
			t.session.Commit()

			committedOffset = lastOffset
			hasCommitted = true

			for _, committed := range queue[start:head] {
				for _, waiter := range committed.waiters {
					waiter <- nil
				}
				committed.waiters = nil
			}

			if head >= 64 && head*2 >= len(queue) {
				n := copy(queue, queue[head:])
				queue = queue[:n]
				head = 0
			}
		}
	}
}
