package handler

import (
	"context"
	"fmt"

	userv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/user/v1"
)

func (h *Handler) ValidateSession(
	ctx context.Context,
	req *userv1.ValidateSessionRequest,
) (*userv1.ValidateSessionResponse, error) {
	claims, err := h.usersService.ValidateSession(ctx, req.AccessToken)
	if err != nil {
		return nil, mapDomainError(fmt.Errorf("failed to validate session: %w", err))
	}

	return &userv1.ValidateSessionResponse{
		UserId:    claims.UserID.String(),
		SessionId: claims.SessionID.String(),
	}, nil
}
