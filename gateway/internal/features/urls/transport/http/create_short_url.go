package urls_transport_http

import (
	"fmt"
	"net/http"
	"strings"

	core_auth "github.com/dyingvoid/shorturl/gateway/internal/core/auth"
	core_errors "github.com/dyingvoid/shorturl/gateway/internal/core/errors"
	core_http_request "github.com/dyingvoid/shorturl/gateway/internal/core/transport/http/request"
	core_http_response "github.com/dyingvoid/shorturl/gateway/internal/core/transport/http/response"
	"github.com/dyingvoid/shorturl/shared/pkg/logger"
	urlv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/url/v1"
)

type CreateShortURLRequest struct {
	OriginalURL string `json:"original_url"`
	ExpiresIn   *int64 `json:"expires_in"`
}

func (r CreateShortURLRequest) Validate() error {
	if strings.TrimSpace(r.OriginalURL) == "" {
		return fmt.Errorf("original_url is required")
	}
	if r.ExpiresIn != nil && *r.ExpiresIn <= 0 {
		return fmt.Errorf("expires_in must be positive")
	}

	return nil
}

type CreateShortURLResponse struct {
	ShortCode string `json:"short_code"`
	ShortURL  string `json:"short_url"`
}

func (h *URLsHTTPHandler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	response := core_http_response.NewHandler(log, w)

	var req CreateShortURLRequest
	if err := core_http_request.DecodeAndValidate(r, &req); err != nil {
		response.Error(err, "failed to decode and validate HTTP request")
		return
	}

	result, err := h.urls.CreateShortURL(ctx, &urlv1.CreateShortURLRequest{
		OriginalUrl: req.OriginalURL,
		UserId:      core_auth.UserID(ctx),
		ExpiresIn:   req.ExpiresIn,
	})
	if err != nil {
		response.Error(core_errors.FromGRPC(err), "failed to create short URL")
		return
	}

	response.JSON(CreateShortURLResponse{
		ShortCode: result.GetShortCode(),
		ShortURL:  result.GetShortUrl(),
	}, http.StatusCreated)
}
