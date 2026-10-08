package handler

import (
	"context"
	"fmt"

	urlv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/url/v1"
)

func (h *Handler) DeleteURL(ctx context.Context, r *urlv1.DeleteURLRequest) (*urlv1.DeleteURLResponse, error) {
	if err := h.urlService.DeleteURL(ctx, r.ShortCode, r.UserId); err != nil {
		return nil, mapDomainError(fmt.Errorf("failed to delete url: %w", err))
	}

	return &urlv1.DeleteURLResponse{}, nil
}
