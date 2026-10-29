package analytics_transport_http

import (
	"context"
	"net/http"

	core_http_middleware "github.com/dyingvoid/shorturl/gateway/internal/core/transport/http/middleware"
	core_http_server "github.com/dyingvoid/shorturl/gateway/internal/core/transport/http/server"
	analyticsv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/analytics/v1"
	"google.golang.org/grpc"
)

type AnalyticsService interface {
	GetAnalytics(ctx context.Context, in *analyticsv1.GetAnalyticsRequest, opts ...grpc.CallOption) (*analyticsv1.GetAnalyticsResponse, error)
}

type AnalyticsHTTPHandler struct {
	analytics AnalyticsService
	auth      core_http_middleware.Middleware
}

func NewAnalyticsHTTPHandler(analytics AnalyticsService, auth core_http_middleware.Middleware) *AnalyticsHTTPHandler {
	return &AnalyticsHTTPHandler{
		analytics: analytics,
		auth:      auth,
	}
}

func (h *AnalyticsHTTPHandler) Routes() []core_http_server.Route {
	return []core_http_server.Route{
		core_http_server.NewRoute(http.MethodGet, "/analytics/{code}", h.GetAnalytics, h.auth),
	}
}
