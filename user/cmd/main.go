package main

import (
	"context"
	"os/signal"
	"syscall"

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
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(), syscall.SIGINT, syscall.SIGTERM,
	)
	defer stop()

	cfg := config.NewMust()

	log := logger.Init(logger.NewMust())
	log.Warn("application started")
	log.Info("time zone", "timezone", cfg.TimeZone.String())

	pool, err := postgres.Connect(ctx, postgres.NewMust())
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	redisClient, err := shared_redis.New(ctx, redis.NewMust())
	if err != nil {
		panic(err)
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

	s := server.New(cfg, log)
	userv1.RegisterUserServiceServer(s.GRPC(), handler.New(usersService))

	if err := s.Run(ctx); err != nil {
		log.Error(err.Error())
	}
}
