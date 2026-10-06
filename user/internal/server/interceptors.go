package server

import (
	"context"
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/dyingvoid/shorturl/user/internal/infrastructure/logger"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const requestIDKey = "x-request-id"

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

func RequestID() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		requestID := requestIDFromContext(ctx)
		if requestID == "" {
			requestID = uuid.NewString()
		}

		md := metadata.Pairs(requestIDKey, requestID)
		if incoming, ok := metadata.FromIncomingContext(ctx); ok {
			md = metadata.Join(incoming, md)
		}

		ctx = metadata.NewIncomingContext(ctx, md)

		_ = grpc.SetHeader(ctx, metadata.Pairs(requestIDKey, requestID))

		return handler(ctx, req)
	}
}

func Logger(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		l := log.With(
			"request_id", requestIDFromContext(ctx),
			"method", info.FullMethod,
		)

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

func requestIDFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	values := md.Get(requestIDKey)
	if len(values) == 0 {
		return ""
	}

	return values[0]
}
