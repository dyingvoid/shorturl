package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dyingvoid/urlshort/url/internal/domain"
	domain_errors "github.com/dyingvoid/urlshort/url/internal/domain/errors"
)

func (s *Service) CreateShortURL(
	ctx context.Context,
	userID string,
	originalURL string,
	expiresIn *time.Duration,
) (*domain.Link, error) {
	if err := validateLink(originalURL); err != nil {
		return nil, err
	}

	urlLimit, err := s.usersService.GetLimit(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("url limit unavailable: %w: %w", err, domain_errors.ErrUnavailable)
	}

	linkCounter, err := s.cache.GetLinkCounter(ctx, userID)
	if err != nil {
		if !errors.Is(err, domain_errors.ErrNotFound) {
			return nil, err
		}

		linkCounter, err = s.links.Count(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to count links: %w", err)
		}

		if err := s.cache.SetLinkCounter(ctx, userID, linkCounter); err != nil {
			return nil, err
		}
	}

	linkCounter, err = s.cache.IncLinkCounter(ctx, userID)
	if err != nil {
		return nil, err
	}

	if urlLimit < linkCounter {
		_, _ = s.cache.DecLinkCounter(ctx, userID)
		return nil, fmt.Errorf("reached limit of %d: %w", urlLimit, domain_errors.ErrResourceExhausted)
	}

	// TODO: to app config
	limit := 3
	for range limit {
		link := domain.NewLink(userID, originalURL, expiresIn)
		if linkCounter, err := s.links.Insert(ctx, &link); err != nil {
			switch {
			case errors.Is(err, domain_errors.ErrResourceExhausted):
				_ = s.cache.SetLinkCounter(ctx, userID, linkCounter)
				return nil, err
			case errors.Is(err, domain_errors.ErrUniqueViolation):
				continue
			}
		}
	}

	return nil, fmt.Errorf("failed to create link: %w", domain_errors.ErrUniqueViolation)
}

func validateLink(url string) error {
	return domain_errors.ErrInvalidArgument
}
