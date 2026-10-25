package urls_transport_http

import (
	"fmt"
	"net/http"

	core_auth "github.com/dyingvoid/shorturl/gateway/internal/core/auth"
	core_errors "github.com/dyingvoid/shorturl/gateway/internal/core/errors"
	core_http_request "github.com/dyingvoid/shorturl/gateway/internal/core/transport/http/request"
	core_http_response "github.com/dyingvoid/shorturl/gateway/internal/core/transport/http/response"
	"github.com/dyingvoid/shorturl/shared/pkg/logger"
	urlv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/url/v1"
)

type ListUserURLsResponse struct {
	URLs       []URLItem `json:"urls"`
	NextCursor string    `json:"next_cursor"`
}

type URLItem struct {
	ShortCode   string `json:"short_code"`
	OriginalURL string `json:"original_url"`
	CreatedAt   int64  `json:"created_at"`
	ExpiresAt   int64  `json:"expires_at,omitempty"`
	IsActive    bool   `json:"is_active"`
}

// ListUserURLs lists the authenticated user's short URLs.
// @Summary List user URLs
// @Description Returns a cursor-paginated list of URLs owned by the authenticated user.
// @Tags urls
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Maximum number of URLs to return"
// @Param cursor query string false "Pagination cursor"
// @Success 200 {object} ListUserURLsResponse
// @Failure 400 {object} core_http_response.ErrorResponse
// @Failure 401 {object} core_http_response.ErrorResponse
// @Failure 500 {object} core_http_response.ErrorResponse
// @Router /urls [get]
func (h *URLsHTTPHandler) ListUserURLs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	response := core_http_response.NewHandler(log, w)

	limit, err := core_http_request.QueryInt32(r, "limit")
	if err != nil {
		response.Error(err, "failed to parse query parameters")
		return
	}
	if limit != nil && *limit < 0 {
		response.Error(
			fmt.Errorf("limit must not be negative: %w", core_errors.ErrInvalidArgument),
			"failed to parse query parameters",
		)
		return
	}

	var limitValue int32
	if limit != nil {
		limitValue = *limit
	}

	result, err := h.urls.ListUserURLs(ctx, &urlv1.ListUserURLsRequest{
		UserId: core_auth.UserID(ctx),
		Limit:  limitValue,
		Cursor: core_http_request.QueryString(r, "cursor"),
	})
	if err != nil {
		response.Error(core_errors.FromGRPC(err), "failed to list user URLs")
		return
	}

	urls := make([]URLItem, 0, len(result.GetUrls()))
	for _, link := range result.GetUrls() {
		urls = append(urls, URLItem{
			ShortCode:   link.GetShortCode(),
			OriginalURL: link.GetOriginalUrl(),
			CreatedAt:   link.GetCreatedAt(),
			ExpiresAt:   link.GetExpiresAt(),
			IsActive:    link.GetIsActive(),
		})
	}

	response.JSON(ListUserURLsResponse{
		URLs:       urls,
		NextCursor: result.GetNextCursor(),
	}, http.StatusOK)
}
