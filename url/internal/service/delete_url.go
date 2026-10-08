package service

import (
	"context"
	"fmt"

	domain_errors "github.com/dyingvoid/urlshort/url/internal/domain/errors"
)

func (s *Service) DeleteURL(
	ctx context.Context,
	shortCode string,
	userID string,
) error {
	if shortCode == "" {
		return fmt.Errorf("empty short code: %w", domain_errors.ErrInvalidArgument)
	}

	link, err := s.links.GetByShortCode(ctx, shortCode)
	if err != nil {
		return fmt.Errorf("failed to get link: %w", err)
	}

	if link.UserID != userID {
		return fmt.Errorf(
			"user '%s' is not the owner of link '%s': %w",
			userID,
			shortCode,
			domain_errors.ErrPermissionDenied,
		)
	}

	if err := s.links.Delete(ctx, shortCode); err != nil {
		return fmt.Errorf("failed to delete link: %w", err)
	}

	_ = s.cache.DeleteLink(ctx, shortCode)
	_, _ = s.cache.DecLinkCounter(ctx, userID)

	return nil
}
