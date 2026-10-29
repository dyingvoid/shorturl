package auth_transport_http

import (
	"fmt"
	"net/http"
	"strings"

	core_errors "github.com/dyingvoid/shorturl/gateway/internal/core/errors"
	core_http_request "github.com/dyingvoid/shorturl/gateway/internal/core/transport/http/request"
	core_http_response "github.com/dyingvoid/shorturl/gateway/internal/core/transport/http/response"
	"github.com/dyingvoid/shorturl/shared/pkg/logger"
	userv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/user/v1"
)

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (r RefreshTokenRequest) Validate() error {
	if strings.TrimSpace(r.RefreshToken) == "" {
		return fmt.Errorf("refresh_token is required")
	}

	return nil
}

type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
}

// RefreshToken issues a new access token.
// @Summary Refresh access token
// @Description Exchanges a valid refresh token for a new access token.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token"
// @Success 200 {object} RefreshTokenResponse
// @Failure 400 {object} core_http_response.ErrorResponse
// @Failure 401 {object} core_http_response.ErrorResponse
// @Failure 500 {object} core_http_response.ErrorResponse
// @Router /api/v1/auth/refresh [post]
func (h *AuthHTTPHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	response := core_http_response.NewHandler(log, w)

	var req RefreshTokenRequest
	if err := core_http_request.DecodeAndValidate(r, &req); err != nil {
		response.Error(err, "failed to decode and validate HTTP request")
		return
	}

	result, err := h.users.RefreshToken(ctx, &userv1.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		response.Error(core_errors.FromGRPC(err), "failed to refresh token")
		return
	}

	response.JSON(RefreshTokenResponse{AccessToken: result.GetAccessToken()}, http.StatusOK)
}
