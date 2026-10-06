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

	_ = s.cache.Delete(ctx, s.sessionKey(sessionID))

	return nil
}
