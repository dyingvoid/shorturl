package service

import (
	"context"
	"fmt"

	"github.com/dyingvoid/shorturl/user/internal/domain"
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
