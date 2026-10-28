package core

import (
	"context"
)

type Delivery[T any] struct {
	Value T
	Ack   func(context.Context) error
}

type Fetcher[T any] interface {
	Fetch(ctx context.Context, limit int) ([]Delivery[T], error)
}

type Writer[T any] interface {
	WriteBatch(ctx context.Context, events []T) error
}
