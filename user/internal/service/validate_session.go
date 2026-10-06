package service

import (
	"context"
	"fmt"

	"github.com/dyingvoid/shorturl/user/internal/domain"
	domain_errors "github.com/dyingvoid/shorturl/user/internal/domain/errors"
	"github.com/google/uuid"
)

func (s *UsersService) ValidateSession(
	ctx context.Context,
	accessToken string,
) (*domain.TokenClaims, error) {
	claims, err := s.tokenService.ParseAccess(accessToken)
	if err != nil {
		return nil, fmt.Errorf("parse access token: %w", err)
	}

	if err := s.ensureSessionActive(ctx, claims.SessionID); err != nil {
		return nil, err
	}

	return &claims, nil
}

func (s *UsersService) ensureSessionActive(
	ctx context.Context,
	sessionID uuid.UUID,
) error {
	key := s.sessionKey(sessionID)

	cached, err := s.cache.Get(ctx, key)
	if err == nil {
		_, parseErr := uuid.Parse(string(cached))
		if parseErr == nil {
			return nil
		}

		s.logger.Error(
			"parse cached session",
			"error", parseErr,
			"session_id", sessionID,
		)
	}

	session, err := s.sessionsRepository.GetActiveSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf(
			"get active session: %v: %w", err, domain_errors.ErrUnauthenticated,
		)
	}

	_ = s.cache.Set(ctx, key, session.UserID.String(), s.tokenService.AccessTTL())

	return nil
}
