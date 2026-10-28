package application

import (
	"context"

	"github.com/dyingvoid/shorturl/analytics/internal/domain"
)

type Record struct {
	Topic     string
	Partition int
	Offset    int64
	Key       []byte
	Value     []byte
}

type RecordFetcher interface {
	FetchMessage(ctx context.Context) (Record, error)
	CommitMessages(ctx context.Context, records ...Record) error
}

type BatchWriter interface {
	WriteBatch(ctx context.Context, events []domain.ClickEvent) error
}
