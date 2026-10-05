package handler

import (
	"context"
	"fmt"

	userv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/user/v1"
	"github.com/dyingvoid/shorturl/user/internal/domain"
)

func (h *Handler) Login(ctx context.Context, req *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	tokens, err := h.usersService.Login(
		ctx, domain.NewUserCredentials(req.Email, req.Password),
	)
	if err != nil {
		return nil, mapDomainError(fmt.Errorf("failed to login: %w", err))
	}

	return &userv1.LoginResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}
