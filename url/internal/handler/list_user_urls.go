package handler

import (
	"context"
	"fmt"

	urlv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/url/v1"
)

func (h *Handler) ListUserURLs(ctx context.Context, r *urlv1.ListUserURLsRequest) (*urlv1.ListUserURLsResponse, error) {
	links, nextCursor, err := h.urlService.ListUserURLs(ctx, r.UserId, int(r.Limit), r.Cursor)
	if err != nil {
		return nil, mapDomainError(fmt.Errorf("failed to list user urls: %w", err))
	}

	urls := make([]*urlv1.URLItem, 0, len(links))
	for _, link := range links {
		item := &urlv1.URLItem{
			ShortCode:   link.ShortCode,
			OriginalUrl: link.OriginalURL,
			CreatedAt:   link.CreatedAt.Unix(),
			IsActive:    link.IsActive,
		}

		if link.ExpiresAt != nil {
			item.ExpiresAt = link.ExpiresAt.Unix()
		}

		urls = append(urls, item)
	}

	return &urlv1.ListUserURLsResponse{
		Urls:       urls,
		NextCursor: nextCursor,
	}, nil
}
