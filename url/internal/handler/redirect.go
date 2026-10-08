package handler

import (
	"context"
	"fmt"

	urlv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/url/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func (h *Handler) Redirect(ctx context.Context, r *urlv1.RedirectRequest) (*urlv1.RedirectResponse, error) {
	res, err := h.urlService.Redirect(ctx, r.ShortCode)
	if err != nil {
		return nil, mapDomainError(fmt.Errorf("failed to redirect: %w", err))
	}

	_ = grpc.SetHeader(ctx, metadata.Pairs(cacheSourceHeaderKey, res.Origin.String()))

	return &urlv1.RedirectResponse{
		OriginalUrl: res.Value,
	}, nil
}
