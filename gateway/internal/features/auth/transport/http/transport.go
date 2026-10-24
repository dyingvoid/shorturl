package auth_transport_http

import (
	"context"
	"net/http"

	core_http_middleware "github.com/dyingvoid/shorturl/gateway/internal/core/transport/http/middleware"
	core_http_server "github.com/dyingvoid/shorturl/gateway/internal/core/transport/http/server"
	userv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/user/v1"
	"google.golang.org/grpc"
)

type AuthService interface {
	Register(ctx context.Context, in *userv1.RegisterRequest, opts ...grpc.CallOption) (*userv1.RegisterResponse, error)
	Login(ctx context.Context, in *userv1.LoginRequest, opts ...grpc.CallOption) (*userv1.LoginResponse, error)
	Logout(ctx context.Context, in *userv1.LogoutRequest, opts ...grpc.CallOption) (*userv1.LogoutResponse, error)
	ValidateSession(ctx context.Context, in *userv1.ValidateSessionRequest, opts ...grpc.CallOption) (*userv1.ValidateSessionResponse, error)
	RefreshToken(ctx context.Context, in *userv1.RefreshTokenRequest, opts ...grpc.CallOption) (*userv1.RefreshTokenResponse, error)
}

type AuthHTTPHandler struct {
	users AuthService
}

func NewAuthHTTPHandler(users AuthService) *AuthHTTPHandler {
	return &AuthHTTPHandler{users: users}
}

func (h *AuthHTTPHandler) Routes() []core_http_server.Route {
	return []core_http_server.Route{
		{
			Method:  http.MethodPost,
			Path:    "/auth/register",
			Handler: h.Register,
		},
		{
			Method:  http.MethodPost,
			Path:    "/auth/login",
			Handler: h.Login,
		},
		{
			Method:  http.MethodPost,
			Path:    "/auth/refresh",
			Handler: h.RefreshToken,
		},
		{
			Method:     http.MethodPost,
			Path:       "/auth/logout",
			Handler:    h.Logout,
			Middleware: []core_http_middleware.Middleware{h.AuthMiddleware},
		},
	}
}
