package service

import (
	"context"
	"fmt"

	"github.com/dyingvoid/shorturl/shared/pkg/types"
	domain_errors "github.com/dyingvoid/urlshort/url/internal/domain/errors"
)

func (s *Service) Redirect(
	ctx context.Context, shortCode string,
) (types.Sourced[string], error) {
	if shortCode == "" {
		return types.Sourced[string]{}, fmt.Errorf("empty short code: %w", domain_errors.ErrInvalidArgument)
	}

	link, err := s.cache.GetLink(ctx, shortCode)
	origin := types.OriginCache
	if err != nil {
		origin = types.OriginDatabase

		link, err = s.links.GetByShortCode(ctx, shortCode)
		if err != nil {
			return types.Sourced[string]{}, fmt.Errorf("failed to get link: %w", err)
		}
	}

	if !link.Active() {
		return types.Sourced[string]{}, fmt.Errorf("link '%s' is not active: %w", link.ShortCode, domain_errors.ErrNotFound)
	}

	_ = s.cache.SetLink(ctx, link, s.linkCacheTTL)

	return types.Sourced[string]{Value: link.OriginalURL, Origin: origin}, nil
}
