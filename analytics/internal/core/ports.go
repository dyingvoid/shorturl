package core

import (
	"context"
)

type MsgBatch[T any] struct {
	Messages []T
	Ack      func(context.Context) error
}

type Fetcher[T any] interface {
	Fetch(ctx context.Context) (MsgBatch[T], error)
}

type Writer[T any] interface {
	WriteBatch(ctx context.Context, events []T) error
}

type Metrics interface {
	IncKafkaMessagesConsumed(count int)
	ObserveClickHouseBatchInsertDuration(seconds float64)
}
