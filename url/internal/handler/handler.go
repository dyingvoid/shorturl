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

	CreateURL(link domain.Link) string

	GetURL(
		ctx context.Context,
		shortCode string,
	) (*domain.Link, error)

	ListUserURLs(
		ctx context.Context,
		userID string,
		limit int,
		cursor string,
	) ([]domain.Link, string, error)

	DeleteURL(
		ctx context.Context,
		shortCode string,
		userID string,
	) error
}

func New(urlService URLService) *Handler {
	return &Handler{
		urlService: urlService,
	}
}

func (h *Handler) Redirect(_ context.Context, _ *userv1.RedirectRequest) (*userv1.RedirectResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (h *Handler) mustEmbedUnimplementedURLServiceServer() {
	panic("not implemented") // TODO: Implement
}
