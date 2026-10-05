package handler

import (
	"context"

	userv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/user/v1"
	"github.com/dyingvoid/shorturl/user/internal/domain"
)

type Handler struct {
	userv1.UnimplementedUserServiceServer
	usersService UsersService
}

type UsersService interface {
	Register(
		ctx context.Context,
		credentials domain.UserCredentials,
	) (domain.User, error)
	Login(
		ctx context.Context,
		credentials domain.UserCredentials,
	) (domain.TokenPair, error)
}

func NewHandler(usersService UsersService) *Handler {
	return &Handler{
		usersService: usersService,
	}
}

func (h *Handler) Logout(_ context.Context, _ *userv1.LogoutRequest) (*userv1.LogoutResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (h *Handler) ValidateSession(_ context.Context, _ *userv1.ValidateSessionRequest) (*userv1.ValidateSessionResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (h *Handler) RefreshToken(_ context.Context, _ *userv1.RefreshTokenRequest) (*userv1.RefreshTokenResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (h *Handler) GetLimit(_ context.Context, _ *userv1.GetLimitRequest) (*userv1.GetLimitResponse, error) {
	panic("not implemented") // TODO: Implement
}
