package service

import (
	"context"
	"fmt"

	"github.com/dyingvoid/urlshort/url/internal/domain"
)

const (
	defaultListLimit = 20
	maxListLimit     = 100
)

func (s *Service) ListUserURLs(
	ctx context.Context,
	userID string,
	limit int,
	cursor string,
) ([]domain.Link, string, error) {
	limit = normalizeListLimit(limit)

	links, err := s.links.ListByUser(ctx, userID, limit+1, cursor)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list links: %w", err)
	}

	if len(links) <= limit {
		return links, "", nil
	}

	links = links[:limit]

	return links, links[limit-1].ID, nil
}

func normalizeListLimit(limit int) int {
	switch {
	case limit <= 0:
		return defaultListLimit
	case limit > maxListLimit:
		return maxListLimit
	default:
		return limit
	}
}
