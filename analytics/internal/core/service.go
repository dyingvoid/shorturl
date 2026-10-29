package core

import (
	"context"
	"time"

	"github.com/dyingvoid/shorturl/analytics/internal/domain"
)

type Service struct {
	cfg       Config
	urls      URLService
	analytics AnalyticsRepository
}

func NewService(cfg Config, urls URLService, analytics AnalyticsRepository) *Service {
	return &Service{
		cfg:       cfg,
		urls:      urls,
		analytics: analytics,
	}
}

type URLService interface {
	GetURL(
		ctx context.Context,
		shortCode string,
	) (domain.ShortURL, error)
}

type AnalyticsRepository interface {
	TotalClicks(
		ctx context.Context,
		shortCode string,
	) (int64, error)
	ClicksByDay(
		ctx context.Context,
		shortCode string,
		from time.Time,
	) ([]domain.DayClicks, error)
	TopReferers(
		ctx context.Context,
		shortCode string,
		from time.Time,
		limit int,
	) ([]domain.RefererClicks, error)
}
