package urls_transport_http

import (
	"context"
	"net/http"

	core_http_middleware "github.com/dyingvoid/shorturl/gateway/internal/core/transport/http/middleware"
	core_http_server "github.com/dyingvoid/shorturl/gateway/internal/core/transport/http/server"
	urlv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/url/v1"
	"google.golang.org/grpc"
)

type URLsService interface {
	CreateShortURL(ctx context.Context, in *urlv1.CreateShortURLRequest, opts ...grpc.CallOption) (*urlv1.CreateShortURLResponse, error)
	DeleteURL(ctx context.Context, in *urlv1.DeleteURLRequest, opts ...grpc.CallOption) (*urlv1.DeleteURLResponse, error)
	ListUserURLs(ctx context.Context, in *urlv1.ListUserURLsRequest, opts ...grpc.CallOption) (*urlv1.ListUserURLsResponse, error)
	Redirect(ctx context.Context, in *urlv1.RedirectRequest, opts ...grpc.CallOption) (*urlv1.RedirectResponse, error)
}

type URLsHTTPHandler struct {
	urls URLsService
	auth core_http_middleware.Middleware
}

func NewURLsHTTPHandler(urls URLsService, auth core_http_middleware.Middleware) *URLsHTTPHandler {
	return &URLsHTTPHandler{
		urls: urls,
		auth: auth,
	}
}

func (h *URLsHTTPHandler) Routes() []core_http_server.Route {
	middleware := []core_http_middleware.Middleware{h.auth}

	return []core_http_server.Route{
		{
			Method:     http.MethodPost,
			Path:       "/urls",
			Handler:    h.CreateShortURL,
			Middleware: middleware,
		},
		{
			Method:     http.MethodGet,
			Path:       "/urls",
			Handler:    h.ListUserURLs,
			Middleware: middleware,
		},
		{
			Method:     http.MethodDelete,
			Path:       "/urls/{code}",
			Handler:    h.DeleteURL,
			Middleware: middleware,
		},
	}
}

func (h *URLsHTTPHandler) PublicRoutes() []core_http_server.Route {
	return []core_http_server.Route{
		{
			Method:  http.MethodGet,
			Path:    "/{code}",
			Handler: h.Redirect,
		},
	}
}
