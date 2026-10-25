package core_http_server

import (
	"net/http"

	core_http_middleware "github.com/dyingvoid/shorturl/gateway/internal/core/transport/http/middleware"
)

type Route struct {
	Method     string
	Path       string
	Handler    http.HandlerFunc
	Middleware []core_http_middleware.Middleware
}

func NewRoute(
	method, path string,
	handler http.HandlerFunc,
	middleware ...core_http_middleware.Middleware,
) Route {
	return Route{
		Method:     method,
		Path:       path,
		Handler:    handler,
		Middleware: middleware,
	}
}

func (r Route) Pattern() string {
	if r.Method == "" {
		return r.Path
	}
	return r.Method + " " + r.Path
}

func (r Route) WithMiddleware() http.Handler {
	return core_http_middleware.Chain(r.Handler, r.Middleware...)
}
