package tracing

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"google.golang.org/grpc/stats"
)

func GRPCServerHandler() stats.Handler {
	return otelgrpc.NewServerHandler()
}

func GRPCClientHandler() stats.Handler {
	return otelgrpc.NewClientHandler()
}

func HTTPMiddleware(operation string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return otelhttp.NewHandler(
			next,
			operation,
			otelhttp.WithSpanNameFormatter(
				func(_ string, r *http.Request) string {
					return r.Method + " " + r.URL.Path
				},
			),
		)
	}
}
