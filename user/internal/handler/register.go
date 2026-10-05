package handler

import (
	"context"

	userv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/user/v1"
	"github.com/dyingvoid/shorturl/user/internal/domain"
)

func (h *Handler) Register(ctx context.Context, req *userv1.RegisterRequest) (*userv1.RegisterResponse, error) {
	user, err := h.usersService.Register(
		ctx, domain.NewUserCredentials(req.Email, req.Password),
	)
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &userv1.RegisterResponse{
		UserId: user.ID.String(),
	}, nil
}
