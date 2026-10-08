package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
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

	linkCounter, err := s.updateLinkCounter(ctx, userID)
	if err != nil {
		return nil, err
	}

	committed := false
	defer func (){
		if !committed {
			_, _ = s.cache.DecLinkCounter(ctx, userID)
		}
	}()

	if urlLimit < linkCounter {
		return nil, fmt.Errorf("reached limit of %d: %w", urlLimit, domain_errors.ErrResourceExhausted)
	}

	for range s.createURLRetries {
		link, err := domain.NewLink(userID, originalURL, expiresIn)
		if err != nil {
			return nil, fmt.Errorf("error creating link: %w", err)
		}

		err = s.links.Insert(ctx, &link)
		switch {
		case err == nil:
			committed = true
			return &link, nil
		case errors.Is(err, domain_errors.ErrUniqueViolation):
			continue
		default:
			return nil, err
		}
	}

	return nil, fmt.Errorf("failed to create link: %w", domain_errors.ErrUniqueViolation)
}

func (s *Service) updateLinkCounter(
	ctx context.Context,
	userID string,
) (int, error) {
	linkCounter, err := s.cache.GetLinkCounter(ctx, userID)
	if err != nil {
		if !errors.Is(err, domain_errors.ErrNotFound) {
			return 0, err
		}

		linkCounter, err = s.links.Count(ctx, userID)
		if err != nil {
			return 0, fmt.Errorf("failed to count links: %w", err)
		}

		if err := s.cache.SetLinkCounter(ctx, userID, linkCounter); err != nil {
			return 0, err
		}
	}

	linkCounter, err = s.cache.IncLinkCounter(ctx, userID)
	if err != nil {
		return 0, err
	}

	return linkCounter, nil
}

func validateLink(rawURL string) error {
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return fmt.Errorf("parse url %q: %w", rawURL, domain_errors.ErrInvalidArgument)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("unsupported url scheme %q: %w", parsed.Scheme, domain_errors.ErrInvalidArgument)
	}

	if parsed.Host == "" {
		return fmt.Errorf("url %q has no host: %w", rawURL, domain_errors.ErrInvalidArgument)
	}

	return nil
}
