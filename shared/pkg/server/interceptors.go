package server

import (
	"context"
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/dyingvoid/shorturl/shared/pkg/logger"
	"github.com/dyingvoid/shorturl/shared/pkg/requestid"
	"github.com/dyingvoid/shorturl/shared/pkg/tracing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RequestID() grpc.UnaryServerInterceptor {
	return requestid.UnaryServerInterceptor()
}

func Logger(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		l := log.With(
			// TODO: perhaps get rid of request_id as there is trace_id
			"request_id", requestIDFromContext(ctx),
			"method", info.FullMethod,
		)

		if traceID := tracing.TraceID(ctx); traceID != "" {
			l = l.With("trace_id", traceID)
		}

		ctx = logger.ToContext(ctx, l)

		return handler(ctx, req)
	}
}

func Trace() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		log := logger.FromContext(ctx)

		before := time.Now()
		log.Debug(
			">>> incoming gRPC request",
			"method", info.FullMethod,
			"time", before.UTC(),
		)

		resp, err := handler(ctx, req)

		log.Debug(
			"<<< done gRPC request",
			"code", status.Code(err).String(),
			"latency", time.Since(before),
		)

		return resp, err
	}
}

func Recovery(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		defer func() {
			if p := recover(); p != nil {
				log.Error(
					"panic",
					"error", p,
					"stack", string(debug.Stack()),
					"method", info.FullMethod,
					"request_id", requestIDFromContext(ctx),
				)

				err = status.Error(codes.Internal, "internal error")
			}
		}()

		return handler(ctx, req)
	}
}

func requestIDFromContext(ctx context.Context) string {
	if id := requestid.FromContext(ctx); id != "" {
		return id
	}

	return requestid.FromIncomingContext(ctx)
}
