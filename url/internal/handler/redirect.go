package handler

import (
	"context"
	"fmt"

	urlv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/url/v1"
)

func (h *Handler) Redirect(ctx context.Context, r *urlv1.RedirectRequest) (*urlv1.RedirectResponse, error) {
	originalURL, err := h.urlService.Redirect(ctx, r.ShortCode)
	if err != nil {
		return nil, mapDomainError(fmt.Errorf("failed to redirect: %w", err))
	}

	return &urlv1.RedirectResponse{
		OriginalUrl: originalURL,
	}, nil
}
