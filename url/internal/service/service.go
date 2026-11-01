package service

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/dyingvoid/urlshort/url/internal/domain"
)

type Service struct {
	url              string
	createURLRetries int
	linkCacheTTL     time.Duration

	usersService UsersService
	cache        Cache
	links        LinksRepository
	producer     Producer
	metrics      Metrics
	logger       *slog.Logger
}

func New(
	url string,
	createURLRetries int,
	linkCacheTTL time.Duration,
	usersService UsersService,
	cache Cache,
	links LinksRepository,
	producer Producer,
	metrics Metrics,
	logger *slog.Logger,
) *Service {
	return &Service{
		url:              url,
		createURLRetries: createURLRetries,
		linkCacheTTL:     linkCacheTTL,
		usersService:     usersService,
		cache:            cache,
		links:            links,
		producer:         producer,
		metrics:          metrics,
		logger:           logger,
	}
}

func (s *Service) CreateURL(link domain.Link) string {
	return fmt.Sprintf("%s/%s", s.url, link.ShortCode)
}
