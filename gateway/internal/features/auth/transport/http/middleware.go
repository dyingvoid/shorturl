package auth_transport_http

import (
	"fmt"
	"net/http"
	"strings"

	core_auth "github.com/dyingvoid/shorturl/gateway/internal/core/auth"
	core_errors "github.com/dyingvoid/shorturl/gateway/internal/core/errors"
	core_http_response "github.com/dyingvoid/shorturl/gateway/internal/core/transport/http/response"
	"github.com/dyingvoid/shorturl/shared/pkg/logger"
	userv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/user/v1"
)

func (h *AuthHTTPHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log := logger.FromContext(ctx)
		response := core_http_response.NewHandler(log, w)

		token, err := bearerToken(r)
		if err != nil {
			response.Error(err, "failed to authenticate request")
			return
		}

		result, err := h.users.ValidateSession(ctx, &userv1.ValidateSessionRequest{
			AccessToken: token,
		})
		if err != nil {
			response.Error(core_errors.FromGRPC(err), "failed to validate session")
			return
		}

		identity := core_auth.Identity{
			UserID:    result.GetUserId(),
			SessionID: result.GetSessionId(),
		}
		next.ServeHTTP(w, r.WithContext(core_auth.ToContext(ctx, identity)))
	})
}

func bearerToken(r *http.Request) (string, error) {
	parts := strings.Fields(r.Header.Get("Authorization"))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", fmt.Errorf("missing bearer token: %w", core_errors.ErrUnauthenticated)
	}

	return parts[1], nil
}
