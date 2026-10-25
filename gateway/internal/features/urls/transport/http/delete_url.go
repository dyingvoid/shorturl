package urls_transport_http

import (
	"net/http"

	core_auth "github.com/dyingvoid/shorturl/gateway/internal/core/auth"
	core_errors "github.com/dyingvoid/shorturl/gateway/internal/core/errors"
	core_http_request "github.com/dyingvoid/shorturl/gateway/internal/core/transport/http/request"
	core_http_response "github.com/dyingvoid/shorturl/gateway/internal/core/transport/http/response"
	"github.com/dyingvoid/shorturl/shared/pkg/logger"
	urlv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/url/v1"
)

// DeleteURL deletes a short URL.
// @Summary Delete URL
// @Description Deletes a short URL owned by the authenticated user.
// @Tags urls
// @Produce json
// @Security BearerAuth
// @Param code path string true "Short code"
// @Success 204 "No Content"
// @Failure 400 {object} core_http_response.ErrorResponse
// @Failure 401 {object} core_http_response.ErrorResponse
// @Failure 403 {object} core_http_response.ErrorResponse
// @Failure 404 {object} core_http_response.ErrorResponse
// @Failure 500 {object} core_http_response.ErrorResponse
// @Router /urls/{code} [delete]
func (h *URLsHTTPHandler) DeleteURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	response := core_http_response.NewHandler(log, w)

	code, err := core_http_request.PathValue(r, "code")
	if err != nil {
		response.Error(err, "failed to parse path parameters")
		return
	}

	if _, err := h.urls.DeleteURL(ctx, &urlv1.DeleteURLRequest{
		ShortCode: code,
		UserId:    core_auth.UserID(ctx),
	}); err != nil {
		response.Error(core_errors.FromGRPC(err), "failed to delete URL")
		return
	}

	response.NoContent()
}
