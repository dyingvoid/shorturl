package auth_transport_http

import (
	"net/http"

	core_auth "github.com/dyingvoid/shorturl/gateway/internal/core/auth"
	core_errors "github.com/dyingvoid/shorturl/gateway/internal/core/errors"
	core_http_response "github.com/dyingvoid/shorturl/gateway/internal/core/transport/http/response"
	"github.com/dyingvoid/shorturl/shared/pkg/logger"
	userv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/user/v1"
)

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
