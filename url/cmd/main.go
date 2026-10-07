package main

import (
	"context"
	"os/signal"
	"syscall"

	urlv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/url/v1"

	"github.com/dyingvoid/shorturl/shared/pkg/logger"
	"github.com/dyingvoid/shorturl/shared/pkg/server"
	"github.com/dyingvoid/urlshort/url/internal/config"
	"github.com/dyingvoid/urlshort/url/internal/handler"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(), syscall.SIGINT, syscall.SIGTERM,
	)
	defer stop()

	cfg := config.NewMust()
	log := logger.Init(logger.NewMust())

	s := server.New(cfg, log)
	urlv1.RegisterURLServiceServer(s.GRPC(), handler.New())
	if err := s.Run(ctx); err != nil {
		log.Error(err.Error())
	}
}
