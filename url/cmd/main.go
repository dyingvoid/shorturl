package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/dyingvoid/shorturl/shared/pkg/logger"
	urlv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/url/v1"
	shared_redis "github.com/dyingvoid/shorturl/shared/pkg/redis"
	"github.com/dyingvoid/shorturl/shared/pkg/server"
	"github.com/dyingvoid/urlshort/url/internal/config"
	"github.com/dyingvoid/urlshort/url/internal/handler"
	"github.com/dyingvoid/urlshort/url/internal/infrastructure/mongo"
	links_mongo_repository "github.com/dyingvoid/urlshort/url/internal/infrastructure/mongo/repository"
	"github.com/dyingvoid/urlshort/url/internal/infrastructure/redis"
	"github.com/dyingvoid/urlshort/url/internal/service"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(), syscall.SIGINT, syscall.SIGTERM,
	)
	defer stop()

	cfg := config.NewMust()
	time.Local = cfg.TimeZone
	log := logger.Init(logger.NewMust())

	mongoDB, err := mongo.Connect(ctx, mongo.NewMust())
	if err != nil {
		panic(err)
	}
	defer mongoDB.Client().Disconnect(ctx)

	redisClient, err := shared_redis.New(ctx, redis.NewMust())
	if err != nil {
		panic(err)
	}
	defer redisClient.Close()

	redisCache := redis.NewCache(redisClient, log)
	linksRepository := links_mongo_repository.NewLinksRepository(mongoDB)
	if err := linksRepository.EnsureIndexes(ctx); err != nil {
		panic(err)
	}
	urlService := service.New(nil, redisCache, linksRepository, cfg.CreateURLAttempts)

	s := server.New(cfg, log)
	urlv1.RegisterURLServiceServer(s.GRPC(), handler.New(urlService))
	if err := s.Run(ctx); err != nil {
		log.Error(err.Error())
	}
}
