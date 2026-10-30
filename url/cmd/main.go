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
	infrastructure_kafka "github.com/dyingvoid/urlshort/url/internal/infrastructure/kafka"
	"github.com/dyingvoid/urlshort/url/internal/infrastructure/mongo"
	mongo_repository_links "github.com/dyingvoid/urlshort/url/internal/infrastructure/mongo/links_repository"
	"github.com/dyingvoid/urlshort/url/internal/infrastructure/redis"
	"github.com/dyingvoid/urlshort/url/internal/infrastructure/users_client"
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
	log.Warn("application starting")

	mongoCfg := mongo.NewMust()
	mongoClient, err := mongo.Connect(ctx, mongoCfg)
	if err != nil {
		panic(err)
	}
	defer mongoClient.Disconnect(ctx)

	redisClient, err := shared_redis.New(ctx, redis.NewMust())
	if err != nil {
		panic(err)
	}
	defer redisClient.Close()

	usersClient, err := users_client.Connect(ctx, users_client.NewMust())
	if err != nil {
		panic(err)
	}
	defer usersClient.Close()

	kafkaCfg := infrastructure_kafka.NewMust()
	userViewProducer, err := infrastructure_kafka.NewUserViewProducer(log, kafkaCfg)
	if err != nil {
		panic(err)
	}
	defer userViewProducer.Close()

	redisCache := redis.NewCache(redisClient, log)
	linksRepository := mongo_repository_links.NewLinksRepository(mongoClient.Database(mongoCfg.DB))
	if err := linksRepository.EnsureIndexes(ctx); err != nil {
		panic(err)
	}
	urlService := service.New(
		cfg.URL, cfg.CreateURLAttempts, cfg.LinkCacheTTL,
		usersClient, redisCache, linksRepository, userViewProducer, log,
	)

	// TODO: proper mertrics registry
	s := server.New(cfg, log, "url", nil)
	urlv1.RegisterURLServiceServer(s.GRPC(), handler.New(urlService))

	if err := s.Run(ctx); err != nil {
		log.Error(err.Error())
	}
}
