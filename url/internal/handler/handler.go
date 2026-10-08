package handler

import (
	"context"
	"time"

	urlv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/url/v1"
	userv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/url/v1"
	"github.com/dyingvoid/urlshort/url/internal/domain"
)

// impl --dir ./shared/pkg/proto/url/v1 'h *Handler' 'URLServiceServer'
type Handler struct {
	urlv1.UnimplementedURLServiceServer
	urlService URLService
}

type URLService interface {
	CreateShortURL(
		ctx context.Context,
		userID string,
		originalURL string,
		expiresIn *time.Duration,
	) (*domain.Link, error)
}

func New(urlService URLService) *Handler {
	return &Handler{
		urlService: urlService,
	}
}

func (h *Handler) CreateShortURL(ctx context.Context, r *userv1.CreateShortURLRequest) (*userv1.CreateShortURLResponse, error) {
	_ = time.Unix(*r.ExpiresIn, 0)
	return nil, nil
}

func (h *Handler) GetURL(_ context.Context, _ *userv1.GetURLRequest) (*userv1.GetURLResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (h *Handler) DeleteURL(_ context.Context, _ *userv1.DeleteURLRequest) (*userv1.DeleteURLResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (h *Handler) ListUserURLs(_ context.Context, _ *userv1.ListUserURLsRequest) (*userv1.ListUserURLsResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (h *Handler) Redirect(_ context.Context, _ *userv1.RedirectRequest) (*userv1.RedirectResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (h *Handler) mustEmbedUnimplementedURLServiceServer() {
	panic("not implemented") // TODO: Implement
}
