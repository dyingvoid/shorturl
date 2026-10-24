package auth_transport_http

import (
	"net/http"

	core_errors "github.com/dyingvoid/shorturl/gateway/internal/core/errors"
	core_http_request "github.com/dyingvoid/shorturl/gateway/internal/core/transport/http/request"
	core_http_response "github.com/dyingvoid/shorturl/gateway/internal/core/transport/http/response"
	"github.com/dyingvoid/shorturl/shared/pkg/logger"
	userv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/user/v1"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r RegisterRequest) Validate() error {
	return validateCredentials(r.Email, r.Password)
}

type RegisterResponse struct {
	UserID string `json:"user_id"`
}

func (h *AuthHTTPHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	response := core_http_response.NewHandler(log, w)

	var req RegisterRequest
	if err := core_http_request.DecodeAndValidate(r, &req); err != nil {
		response.Error(err, "failed to decode and validate HTTP request")
		return
	}

	result, err := h.users.Register(ctx, &userv1.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		response.Error(core_errors.FromGRPC(err), "failed to register user")
		return
	}

	response.JSON(RegisterResponse{UserID: result.GetUserId()}, http.StatusCreated)
}
