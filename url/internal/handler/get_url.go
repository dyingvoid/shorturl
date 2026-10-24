package handler

import (
	"context"
	"fmt"

	urlv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/url/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func (h *Handler) GetURL(ctx context.Context, r *urlv1.GetURLRequest) (*urlv1.GetURLResponse, error) {
	res, err := h.urlService.GetURL(ctx, r.ShortCode)
	if err != nil {
		return nil, mapDomainError(fmt.Errorf("failed to get url: %w", err))
	}

	_ = grpc.SetHeader(ctx, metadata.Pairs(cacheSourceHeaderKey, res.Origin.String()))

	resp := &urlv1.GetURLResponse{
		OriginalUrl: res.Value.OriginalURL,
		UserId:      res.Value.UserID,
		CreatedAt:   res.Value.CreatedAt.Unix(),
		IsActive:    res.Value.IsActive,
	}

	if res.Value.ExpiresAt != nil {
		resp.ExpiresAt = res.Value.ExpiresAt.Unix()
	}

	return resp, nil
}
