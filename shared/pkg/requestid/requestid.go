package requestid

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	Header      = "X-Request-ID"
	MetadataKey = "x-request-id"
)

type contextKey struct{}

var key = contextKey{}

func ToContext(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, key, id)
}

func FromContext(ctx context.Context) string {
	id, _ := ctx.Value(key).(string)
	return id
}

func FromIncomingContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	values := md.Get(MetadataKey)
	if len(values) == 0 {
		return ""
	}

	return values[0]
}

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		id := FromIncomingContext(ctx)
		if id == "" {
			id = uuid.NewString()
		}

		ctx = ToContext(ctx, id)

		if err := grpc.SetHeader(ctx, metadata.Pairs(MetadataKey, id)); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req any,
		reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		id := FromContext(ctx)
		if id == "" {
			id = uuid.NewString()
			ctx = ToContext(ctx, id)
		}

		md, _ := metadata.FromOutgoingContext(ctx)
		md = md.Copy()
		md.Set(MetadataKey, id)

		ctx = metadata.NewOutgoingContext(ctx, md)

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}