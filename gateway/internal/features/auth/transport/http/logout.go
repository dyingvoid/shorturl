package auth_transport_http

import (
	"net/http"

	core_auth "github.com/dyingvoid/shorturl/gateway/internal/core/auth"
	core_errors "github.com/dyingvoid/shorturl/gateway/internal/core/errors"
	core_http_response "github.com/dyingvoid/shorturl/gateway/internal/core/transport/http/response"
	"github.com/dyingvoid/shorturl/shared/pkg/logger"
	userv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/user/v1"
)

// Logout terminates the current session.
// @Summary Logout user
// @Description Invalidates the session identified by the bearer token.
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 204 "No Content"
// @Failure 401 {object} core_http_response.ErrorResponse
// @Failure 500 {object} core_http_response.ErrorResponse
// @Router /api/v1/auth/logout [post]
func (h *AuthHTTPHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	response := core_http_response.NewHandler(log, w)

	if _, err := h.users.Logout(ctx, &userv1.LogoutRequest{
		SessionId: core_auth.SessionID(ctx),
	}); err != nil {
		response.Error(core_errors.FromGRPC(err), "failed to logout user")
		return
	}

	response.NoContent()
}
