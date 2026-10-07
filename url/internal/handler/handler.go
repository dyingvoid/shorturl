package handler

import (
	"context"

	urlv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/url/v1"
	userv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/url/v1"
)

type Handler struct {
	urlv1.UnimplementedURLServiceServer
}

func New() *Handler {
	return &Handler{}
}

func (h *Handler) CreateShortURL(_ context.Context, _ *userv1.CreateShortURLRequest) (*userv1.CreateShortURLResponse, error) {
	panic("not implemented") // TODO: Implement
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
