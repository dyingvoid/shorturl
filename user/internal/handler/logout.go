package handler

import (
	"context"
	"fmt"

	userv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/user/v1"
	domain_errors "github.com/dyingvoid/shorturl/user/internal/domain/errors"
	"github.com/google/uuid"
)

func (h *Handler) Logout(ctx context.Context, req *userv1.LogoutRequest) (*userv1.LogoutResponse, error) {
	sessionID, err := uuid.Parse(req.SessionId)
	if err != nil {
		return nil, mapDomainError(fmt.Errorf(
			"invalid session id: %v: %w", err, domain_errors.ErrInvalidArgument,
		))
	}

	if err := h.usersService.Logout(ctx, sessionID); err != nil {
		return nil, mapDomainError(fmt.Errorf("failed to logout: %w", err))
	}

	return &userv1.LogoutResponse{}, nil
}
