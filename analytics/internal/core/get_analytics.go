package core

import (
	"context"
	"fmt"
	"time"

	"github.com/dyingvoid/shorturl/analytics/internal/domain"
	domain_errors "github.com/dyingvoid/shorturl/analytics/internal/domain/errors"
	"golang.org/x/sync/errgroup"
)

func (s *Service) GetAnalytics(
	ctx context.Context,
	shortCode string,
	userID string,
	days int32,
) (domain.Analytics, error) {
	if shortCode == "" {
		return domain.Analytics{}, fmt.Errorf("empty short code: %w", domain_errors.ErrInvalidArgument)
	}
	if userID == "" {
		return domain.Analytics{}, fmt.Errorf("empty user id: %w", domain_errors.ErrInvalidArgument)
	}

	if days <= 0 {
		days = s.cfg.DefaultDays
	}
	if days > s.cfg.MaxDays {
		return domain.Analytics{}, fmt.Errorf("days must not exceed %d: %w", s.cfg.MaxDays, domain_errors.ErrInvalidArgument)
	}

	link, err := s.urls.GetURL(ctx, shortCode)
	if err != nil {
		return domain.Analytics{}, fmt.Errorf("failed to get url: %w", err)
	}

	if link.UserID != userID {
		return domain.Analytics{}, fmt.Errorf("url '%s' belongs to another user: %w", shortCode, domain_errors.ErrPermissionDenied)
	}

	from := time.Now().AddDate(0, 0, -int(days))

	var (
		total    int64
		byDay    []domain.DayClicks
		referers []domain.RefererClicks
	)

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		value, err := s.analytics.TotalClicks(gctx, shortCode)
		if err != nil {
			return fmt.Errorf("failed to get total clicks: %w", err)
		}

		total = value
		return nil
	})

	g.Go(func() error {
		value, err := s.analytics.ClicksByDay(gctx, shortCode, from)
		if err != nil {
			return fmt.Errorf("failed to get clicks by day: %w", err)
		}

		byDay = value
		return nil
	})

	g.Go(func() error {
		value, err := s.analytics.TopReferers(gctx, shortCode, from, s.cfg.TopReferers)
		if err != nil {
			return fmt.Errorf("failed to get top referers: %w", err)
		}

		referers = value
		return nil
	})

	if err := g.Wait(); err != nil {
		return domain.Analytics{}, err
	}

	return domain.Analytics{
		ShortCode:   shortCode,
		TotalClicks: total,
		ClicksByDay: byDay,
		TopReferers: referers,
	}, nil
}
