package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dyingvoid/shorturl/shared/pkg/logger"
	userv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/user/v1"
	shared_redis "github.com/dyingvoid/shorturl/shared/pkg/redis"
	"github.com/dyingvoid/shorturl/shared/pkg/server"
	"github.com/dyingvoid/shorturl/user/internal/config"
	"github.com/dyingvoid/shorturl/user/internal/handler"
	"github.com/dyingvoid/shorturl/user/internal/infrastructure/hashing"
	"github.com/dyingvoid/shorturl/user/internal/infrastructure/jwt"
	postgres "github.com/dyingvoid/shorturl/user/internal/infrastructure/postgres/pool"
	users_postgres_repository "github.com/dyingvoid/shorturl/user/internal/infrastructure/postgres/repository"
	"github.com/dyingvoid/shorturl/user/internal/infrastructure/redis"
	"github.com/dyingvoid/shorturl/user/internal/service"
	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	log := logger.Init(logger.NewMust())
	log.Warn("application started")

	if err := run(log); err != nil {
		log.Error("user stopped with error", "err", err)
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
	log.Info("time zone", "timezone", cfg.TimeZone.String())

	pool, err := postgres.Connect(ctx, postgres.NewMust())
	if err != nil {
		return err
	}
	defer pool.Close()

	redisClient, err := shared_redis.New(ctx, redis.NewMust())
	if err != nil {
		return err
	}
	defer redisClient.Close()

	redisCache := redis.NewCache(redisClient, log)

	hasher := hashing.NewPasswordHasher()
	tokenService := jwt.NewTokenService(jwt.NewMust())
	usersRepository := users_postgres_repository.NewUsersRepository(pool)
	sessionsRepository := users_postgres_repository.NewSessionsRepository(pool)
	usersService := service.NewUsersService(
		hasher, tokenService,
		usersRepository, sessionsRepository,
		redisCache, log,
	)

	reg := prometheus.NewRegistry()
	s, err := server.New(ctx, cfg, log, "user", reg)
	if err != nil {
		return err
	}
	userv1.RegisterUserServiceServer(s.GRPC(), handler.New(usersService))

	return s.Run(ctx)
}
