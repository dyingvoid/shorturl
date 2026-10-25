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

func (h *URLsHTTPHandler) Routes() [][]core_http_server.Route {
	res := make([][]core_http_server.Route, 0, 2)
	res = append(res, []core_http_server.Route{
		core_http_server.NewRoute(http.MethodPost, "/urls", h.CreateShortURL, h.auth),
		core_http_server.NewRoute(http.MethodGet, "/urls", h.ListUserURLs, h.auth),
		core_http_server.NewRoute(http.MethodDelete, "/urls/{code}", h.DeleteURL, h.auth),
	})

	// public unversioned
	res = append(res, []core_http_server.Route{
		core_http_server.NewRoute(http.MethodGet, "/{code}", h.Redirect),
	})

	return res
}
