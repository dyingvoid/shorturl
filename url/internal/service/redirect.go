package service

import (
	"context"
	"fmt"

	domain_errors "github.com/dyingvoid/urlshort/url/internal/domain/errors"
)

func (s *Service) Redirect(
	ctx context.Context, shortCode string,
) (string, error) {
	if shortCode == "" {
		return "", fmt.Errorf("empty short code: %w", domain_errors.ErrInvalidArgument)
	}

	link, err := s.cache.GetLink(ctx, shortCode)
	if err != nil {
		link, err = s.links.GetByShortCode(ctx, shortCode)
		if err != nil {
			return "", fmt.Errorf("failed to get link: %w", err)
		}
	}

	if !link.Active() {
		return "", fmt.Errorf("link '%s' is not active: %w", link.ShortCode, domain_errors.ErrNotFound)
	}

	_ = s.cache.SetLink(ctx, link, s.linkCacheTTL)

	return link.OriginalURL, nil
}
