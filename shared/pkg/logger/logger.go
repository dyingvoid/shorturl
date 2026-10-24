package logger

import (
	"log/slog"
	"os"
)

func Init(config Config) *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: config.Level,
	}))
}
