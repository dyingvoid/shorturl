package core_http_server

import (
	"net/http"

	core_http_middleware "github.com/dyingvoid/shorturl/gateway/internal/core/transport/http/middleware"
	"github.com/go-chi/chi/v5"
)

type Version string

const (
	V1 Version = "v1"
)

type VersionRouter struct {
	router     chi.Router
	version    Version
	middleware []core_http_middleware.Middleware
}

func NewVersionRouter(version Version, middleware ...core_http_middleware.Middleware) *VersionRouter {
	return &VersionRouter{
		router:     chi.NewRouter(),
		version:    version,
		middleware: middleware,
	}
}

func (r *VersionRouter) RegisterRoutes(routes ...Route) {
	registerRoutes(r.router, routes...)
}

func (r *VersionRouter) WithMiddleware() http.Handler {
	return core_http_middleware.Chain(r.router, r.middleware...)
}
