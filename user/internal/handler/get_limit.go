package handler

import (
	"context"
	"fmt"

	userv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/user/v1"
	domain_errors "github.com/dyingvoid/shorturl/user/internal/domain/errors"
	"github.com/google/uuid"
)

func (h *Handler) GetLimit(ctx context.Context, req *userv1.GetLimitRequest) (*userv1.GetLimitResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, mapDomainError(fmt.Errorf(
			"invalid user id: %v: %w", err, domain_errors.ErrInvalidArgument,
		))
	}

	limit, err := h.usersService.GetLimit(ctx, userID)
	if err != nil {
		return nil, mapDomainError(fmt.Errorf("failed to get limit: %w", err))
	}

	return &userv1.GetLimitResponse{
		LinksLimit: int32(limit),
	}, nil
}
