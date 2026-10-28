package kafka

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/IBM/sarama"
	"github.com/dyingvoid/shorturl/analytics/internal/application"
)

var _ application.RecordFetcher = (*Reader)(nil)

var errReaderClosed = errors.New("kafka reader is closed")

type commitRequest struct {
	records []application.Record
	result  chan error
}

type Reader struct {
	group     sarama.ConsumerGroup
	topic     string
	handler   *consumerGroupHandler
	messages  chan application.Record
	commits   chan commitRequest
	errors    chan error
	ctx       context.Context
	cancel    context.CancelFunc
	done      chan struct{}
	closeOnce sync.Once
	closeErr  error
}

type consumerGroupHandler struct {
	messages chan<- application.Record
	commits  <-chan commitRequest
}

func NewReader(cfg Config) (*Reader, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.ClientID = "analytics"
	saramaConfig.Version = sarama.V2_1_0_0
	saramaConfig.Consumer.Return.Errors = false
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	saramaConfig.Consumer.Offsets.AutoCommit.Enable = false

	group, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("create consumer group: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	reader := &Reader{
		group:    group,
		topic:    cfg.Topic,
		messages: make(chan application.Record),
		commits:  make(chan commitRequest),
		errors:   make(chan error, 1),
		ctx:      ctx,
		cancel:   cancel,
		done:     make(chan struct{}),
	}
	reader.handler = &consumerGroupHandler{
		messages: reader.messages,
		commits:  reader.commits,
	}

	go reader.consume()

	return reader, nil
}

func (r *Reader) FetchMessage(ctx context.Context) (application.Record, error) {
	select {
	case record := <-r.messages:
		return record, nil
	case err := <-r.errors:
		return application.Record{}, err
	case <-r.done:
		select {
		case err := <-r.errors:
			return application.Record{}, err
		default:
			return application.Record{}, errReaderClosed
		}
	case <-ctx.Done():
		return application.Record{}, ctx.Err()
	}
}

func (r *Reader) CommitMessages(ctx context.Context, records ...application.Record) error {
	if len(records) == 0 {
		return nil
	}

	request := commitRequest{
		records: records,
		result:  make(chan error, 1),
	}

	select {
	case r.commits <- request:
	case <-r.done:
		return errReaderClosed
	case <-ctx.Done():
		return ctx.Err()
	}

	select {
	case err := <-request.result:
		if err != nil {
			return fmt.Errorf("commit messages: %w", err)
		}
		return nil
	case <-r.done:
		return errReaderClosed
	case <-ctx.Done():
		return ctx.Err()
	}
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
		if err != nil {
			if errors.Is(err, sarama.ErrClosedConsumerGroup) || r.ctx.Err() != nil {
				return
			}

			r.reportError(fmt.Errorf("consume kafka messages: %w", err))
			return
		}

		if r.ctx.Err() != nil {
			return
		}
	}
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
	for {
		select {
		case <-session.Context().Done():
			return nil
		case request := <-h.commits:
			request.result <- h.commit(session, request.records)
		case message, ok := <-claim.Messages():
			if !ok {
				return nil
			}

			record := application.Record{
				Topic:     message.Topic,
				Partition: int(message.Partition),
				Offset:    message.Offset,
				Key:       message.Key,
				Value:     message.Value,
			}

			select {
			case h.messages <- record:
			case <-session.Context().Done():
				return nil
			}
		}
	}
}

func (h *consumerGroupHandler) commit(
	session sarama.ConsumerGroupSession,
	records []application.Record,
) error {
	for _, record := range records {
		session.MarkOffset(record.Topic, int32(record.Partition), record.Offset+1, "")
	}

	if err := session.Commit(); err != nil {
		return fmt.Errorf("commit offsets: %w", err)
	}

	return nil
}
