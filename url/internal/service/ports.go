package service

import (
	"context"
	"time"

	"github.com/dyingvoid/urlshort/url/internal/domain"
)

type UsersService interface {
	GetLimit(
		ctx context.Context,
		userID string,
	) (int, error)
}

type Cache interface {
	GetLinkCounter(
		ctx context.Context,
		userID string,
	) (int, error)
	SetIncCounterAtomic(
		ctx context.Context,
		userID string,
		val int,
	) (int, error)
	IncLinkCounter(
		ctx context.Context,
		userID string,
	) (int, error)
	DecLinkCounter(
		ctx context.Context,
		userID string,
	) (int, error)
	GetLink(
		ctx context.Context,
		shortCode string,
	) (*domain.Link, error)
	SetLink(
		ctx context.Context,
		link *domain.Link,
		ttl time.Duration,
	) error
	DeleteLink(
		ctx context.Context,
		shortCode string,
	) error
}

type LinksRepository interface {
	Count(
		ctx context.Context,
		userID string,
	) (int, error)
	Insert(
		ctx context.Context,
		link *domain.Link,
	) error
	GetByShortCode(
		ctx context.Context,
		shortCode string,
	) (*domain.Link, error)
	ListByUser(
		ctx context.Context,
		userID string,
		limit int,
		cursor string,
	) ([]domain.Link, error)
	Delete(
		ctx context.Context,
		shortCode string,
	) error
}

type Producer interface {
	ProduceUserView(domain.UserView) error
}

type Metrics interface {
	IncLinksCreated()
	IncRedirects()
	IncKafkaMessagesProduced()
}
