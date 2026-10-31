package tracing

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
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
		handler := http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			sc := trace.SpanContextFromContext(r.Context())

			if sc.IsValid() {
				w.Header().Set(
					"X-Trace-ID",
					sc.TraceID().String(),
				)
			}

			next.ServeHTTP(w, r)
		})

		return otelhttp.NewHandler(handler, operation)
	}
}