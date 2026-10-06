package service

import (
	"context"
	"fmt"

	"github.com/dyingvoid/shorturl/user/internal/domain"
)

func (s *UsersService) RefreshToken(
	ctx context.Context,
	refreshToken string,
) (domain.TokenPair, error) {
	claims, err := s.tokenService.ParseRefresh(refreshToken)
	if err != nil {
		return domain.TokenPair{}, fmt.Errorf("parse refresh token: %w", err)
	}

	if err := s.ensureSessionActive(ctx, claims.SessionID); err != nil {
		return domain.TokenPair{}, err
	}

	accessToken, err := s.tokenService.RefreshAccess(claims)
	if err != nil {
		return domain.TokenPair{}, fmt.Errorf("refresh access token: %w", err)
	}

	sessionKey := s.sessionKey(claims.SessionID)
	_ = s.cache.Set(ctx, sessionKey, claims.UserID.String(), s.tokenService.AccessTTL())

	return domain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
