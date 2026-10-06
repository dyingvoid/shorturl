package handler

import (
	"context"
	"fmt"

	userv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/user/v1"
)

func (h *Handler) RefreshToken(
	ctx context.Context,
	req *userv1.RefreshTokenRequest,
) (*userv1.RefreshTokenResponse, error) {
	tokens, err := h.usersService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, mapDomainError(fmt.Errorf("failed to refresh token: %w", err))
	}

	return &userv1.RefreshTokenResponse{
		AccessToken: tokens.AccessToken,
	}, nil
}
