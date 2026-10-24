package auth_transport_http

import (
	"net/http"

	core_errors "github.com/dyingvoid/shorturl/gateway/internal/core/errors"
	core_http_request "github.com/dyingvoid/shorturl/gateway/internal/core/transport/http/request"
	core_http_response "github.com/dyingvoid/shorturl/gateway/internal/core/transport/http/response"
	"github.com/dyingvoid/shorturl/shared/pkg/logger"
	userv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/user/v1"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r LoginRequest) Validate() error {
	return validateCredentials(r.Email, r.Password)
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHTTPHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	response := core_http_response.NewHandler(log, w)

	var req LoginRequest
	if err := core_http_request.DecodeAndValidate(r, &req); err != nil {
		response.Error(err, "failed to decode and validate HTTP request")
		return
	}

	result, err := h.users.Login(ctx, &userv1.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		response.Error(core_errors.FromGRPC(err), "failed to login user")
		return
	}

	response.JSON(LoginResponse{
		AccessToken:  result.GetAccessToken(),
		RefreshToken: result.GetRefreshToken(),
	}, http.StatusOK)
}
