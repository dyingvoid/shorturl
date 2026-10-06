package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (s *UsersService) Logout(
	ctx context.Context,
	sessionID uuid.UUID,
) error {
	if err := s.sessionsRepository.RevokeSession(ctx, sessionID); err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}

	if err := s.cache.Delete(ctx, s.sessionKey(sessionID)); err != nil {
		s.logger.Error(
			"delete cached session",
			"error", err,
			"session_id", sessionID,
		)
	}

	return nil
}
