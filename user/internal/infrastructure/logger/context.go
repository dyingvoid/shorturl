package logger

import (
	"context"
	"log/slog"
)

type contextKey struct{}

var loggerContextKey = contextKey{}

func ToContext(ctx context.Context, log *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey, log)
}

func FromContext(ctx context.Context) *slog.Logger {
	log, ok := ctx.Value(loggerContextKey).(*slog.Logger)
	if !ok {
		return slog.Default()
	}

	return log
}
