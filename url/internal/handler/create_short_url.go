package handler

import (
	"context"
	"time"

	userv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/url/v1"
)

func (h *Handler) CreateShortURL(ctx context.Context, r *userv1.CreateShortURLRequest) (*userv1.CreateShortURLResponse, error) {
	var expiresIn *time.Duration
	if r.ExpiresIn != nil {
		t := time.Until(time.Unix(*r.ExpiresIn, 0))
		expiresIn = &t
	}

	link, err := h.urlService.CreateShortURL(
		ctx, r.UserId, r.OriginalUrl, expiresIn,
	)
	if err != nil {
		return nil, mapDomainError(err)
	}

	return &userv1.CreateShortURLResponse{
		ShortCode: link.ShortCode,
		ShortUrl:  h.urlService.CreateURL(*link),
	}, nil
}
