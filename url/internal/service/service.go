package service

import (
	"context"
	"fmt"

	"github.com/dyingvoid/urlshort/url/internal/domain"
)

type Service struct {
	url              string
	createURLRetries int

	usersService UsersService
	cache        Cache
	links        LinksRepository
}

func New(
	url string,
	createURLRetries int,
	usersService UsersService,
	cache Cache,
	links LinksRepository,
) *Service {
	return &Service{
		url:              url,
		createURLRetries: createURLRetries,
		usersService:     usersService,
		cache:            cache,
		links:            links,
	}
}

func (s *Service) CreateURL(link domain.Link) string {
	return fmt.Sprintf("%s/%s", s.url, link.ShortCode)
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
