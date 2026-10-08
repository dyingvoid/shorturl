package handler

import (
	"context"
	"fmt"

	urlv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/url/v1"
)

func (h *Handler) GetURL(ctx context.Context, r *urlv1.GetURLRequest) (*urlv1.GetURLResponse, error) {
	link, err := h.urlService.GetURL(ctx, r.ShortCode)
	if err != nil {
		return nil, mapDomainError(fmt.Errorf("failed to get url: %w", err))
	}

	resp := &urlv1.GetURLResponse{
		OriginalUrl: link.OriginalURL,
		UserId:      link.UserID,
		CreatedAt:   link.CreatedAt.Unix(),
		IsActive:    link.IsActive,
	}

	if link.ExpiresAt != nil {
		resp.ExpiresAt = link.ExpiresAt.Unix()
	}

	return resp, nil
}
