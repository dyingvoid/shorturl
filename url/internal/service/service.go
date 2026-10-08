package service

import (
	"context"

	"github.com/dyingvoid/urlshort/url/internal/domain"
)

type Service struct {
	usersService     UsersService
	cache            Cache
	links            LinksRepository
	createURLRetries int
}

func New(
	usersService UsersService,
	cache Cache,
	links LinksRepository,
	createURLRetries int,
) *Service {
	return &Service{
		usersService:     usersService,
		cache:            cache,
		links:            links,
		createURLRetries: createURLRetries,
	}
}

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
	SetLinkCounter(
		ctx context.Context,
		userID string,
		val int,
	) error
	IncLinkCounter(
		ctx context.Context,
		userID string,
	) (int, error)
	DecLinkCounter(
		ctx context.Context,
		userID string,
	) (int, error)
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
}
