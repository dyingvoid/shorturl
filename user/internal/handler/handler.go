package handler

import (
	"context"

	userv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/user/v1"
	"github.com/dyingvoid/shorturl/user/internal/domain"
	"github.com/google/uuid"
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
	Logout(
		ctx context.Context,
		sessionID uuid.UUID,
	) error
	RefreshToken(
		ctx context.Context,
		refreshToken string,
	) (domain.TokenPair, error)
	ValidateSession(
		ctx context.Context,
		accessToken string,
	) (*domain.TokenClaims, error)
}

func NewHandler(usersService UsersService) *Handler {
	return &Handler{
		usersService: usersService,
	}
}

func (h *Handler) GetLimit(_ context.Context, _ *userv1.GetLimitRequest) (*userv1.GetLimitResponse, error) {
	panic("not implemented") // TODO: Implement
}
