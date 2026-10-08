package service

import (
	"context"
	"fmt"

	"github.com/dyingvoid/urlshort/url/internal/domain"
	domain_errors "github.com/dyingvoid/urlshort/url/internal/domain/errors"
)

func (s *Service) GetURL(
	ctx context.Context, shortCode string,
) (*domain.Link, error) {
	if shortCode == "" {
		return nil, fmt.Errorf("empty short code: %w", domain_errors.ErrInvalidArgument)
	}

	link, err := s.cache.GetLink(ctx, shortCode)
	if err == nil && link.Active() {
		return link, nil
	}

	link, err = s.links.GetByShortCode(ctx, shortCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get link: %w", err)
	}

	if !link.Active() {
		return nil, fmt.Errorf("link '%s' is not active: %w", link.ShortCode, domain_errors.ErrNotFound)
	}

	_ = s.cache.SetLink(ctx, link, s.linkCacheTTL)

	return link, nil
}
