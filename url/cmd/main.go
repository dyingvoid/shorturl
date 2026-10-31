package main

import (
	"context"
	"log/slog"
	"os"
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
	infrastructure_prometheus "github.com/dyingvoid/urlshort/url/internal/infrastructure/prometheus"
	"github.com/dyingvoid/urlshort/url/internal/infrastructure/redis"
	"github.com/dyingvoid/urlshort/url/internal/infrastructure/users_client"
	"github.com/dyingvoid/urlshort/url/internal/service"
	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	log := logger.Init(logger.NewMust())
	log.Warn("application starting")
	if err := run(log); err != nil {
		log.Error("url stopped with error", "err", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	ctx, stop := signal.NotifyContext(
		context.Background(), syscall.SIGINT, syscall.SIGTERM,
	)
	defer stop()

	cfg := config.NewMust()
	time.Local = cfg.TimeZone

	mongoCfg := mongo.NewMust()
	mongoClient, err := mongo.Connect(ctx, mongoCfg)
	if err != nil {
		return err
	}
	defer mongoClient.Disconnect(ctx)

	redisClient, err := shared_redis.New(ctx, redis.NewMust())
	if err != nil {
		return err
	}
	defer redisClient.Close()

	usersClient, err := users_client.Connect(users_client.NewMust())
	if err != nil {
		return err
	}
	defer usersClient.Close()

	kafkaCfg := infrastructure_kafka.NewMust()
	userViewProducer, err := infrastructure_kafka.NewUserViewProducer(log, kafkaCfg)
	if err != nil {
		return err
	}
	defer userViewProducer.Close()

	redisCache := redis.NewCache(redisClient, log)
	linksRepository := mongo_repository_links.NewLinksRepository(mongoClient.Database(mongoCfg.DB))
	if err := linksRepository.EnsureIndexes(ctx); err != nil {
		return err
	}

	registry := prometheus.NewRegistry()
	metrics, err := infrastructure_prometheus.NewMetrics(registry)
	if err != nil {
		return err
	}

	urlService := service.New(
		cfg.URL, cfg.CreateURLAttempts, cfg.LinkCacheTTL,
		usersClient, redisCache, linksRepository, userViewProducer, metrics, log,
	)

	s, err := server.New(ctx, cfg, log, "url", registry)
	if err != nil {
		return err
	}
	urlv1.RegisterURLServiceServer(s.GRPC(), handler.New(urlService))

	return s.Run(ctx)
}
