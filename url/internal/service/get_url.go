package service

import (
	"context"
	"fmt"

	"github.com/dyingvoid/shorturl/shared/pkg/types"
	"github.com/dyingvoid/urlshort/url/internal/domain"
	domain_errors "github.com/dyingvoid/urlshort/url/internal/domain/errors"
)

func (s *Service) GetURL(
	ctx context.Context, shortCode string,
) (types.Sourced[*domain.Link], error) {
	if shortCode == "" {
		return types.Sourced[*domain.Link]{}, fmt.Errorf("empty short code: %w", domain_errors.ErrInvalidArgument)
	}

	link, err := s.cache.GetLink(ctx, shortCode)
	if err == nil && link.Active() {
		return types.FromCache(link), nil
	}

	link, err = s.links.GetByShortCode(ctx, shortCode)
	if err != nil {
		return types.Sourced[*domain.Link]{}, fmt.Errorf("failed to get link: %w", err)
	}

	if !link.Active() {
		return types.Sourced[*domain.Link]{}, fmt.Errorf("link '%s' is not active: %w", link.ShortCode, domain_errors.ErrNotFound)
	}

	_ = s.cache.SetLink(ctx, link, s.linkCacheTTL)

	return types.FromDatabase(link), nil
}
