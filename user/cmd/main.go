package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	userv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/user/v1"
	"github.com/dyingvoid/shorturl/user/internal/config"
	"github.com/dyingvoid/shorturl/user/internal/handler"
	"github.com/dyingvoid/shorturl/user/internal/infrastructure/hashing"
	"github.com/dyingvoid/shorturl/user/internal/infrastructure/jwt"
	"github.com/dyingvoid/shorturl/user/internal/infrastructure/logger"
	"github.com/dyingvoid/shorturl/user/internal/infrastructure/postgres"
	users_postgres_repository "github.com/dyingvoid/shorturl/user/internal/infrastructure/postgres/repository"
	"github.com/dyingvoid/shorturl/user/internal/infrastructure/redis"
	"github.com/dyingvoid/shorturl/user/internal/server"
	"github.com/dyingvoid/shorturl/user/internal/service"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(), syscall.SIGINT, syscall.SIGTERM,
	)
	defer stop()

	appCfg := config.NewMust()

	appLogger := logger.Init(logger.NewMust())
	appLogger.Warn("application started")
	appLogger.Info("time zone", "timezone", appCfg.TimeZone.String())

	pool, err := postgres.Connect(ctx, postgres.NewMust())
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer pool.Close()

	redisCache, err := redis.Connect(ctx, redis.NewMust(), appLogger)
	if err != nil {
		log.Fatalf("connect redis: %v", err)
	}
	defer redisCache.Close()

	hasher := hashing.NewPasswordHasher()
	tokenService := jwt.NewTokenService(jwt.NewMust())
	usersRepository := users_postgres_repository.NewUsersRepository(pool)
	sessionsRepository := users_postgres_repository.NewSessionsRepository(pool)
	usersService := service.NewUsersService(
		hasher, tokenService,
		usersRepository, sessionsRepository,
		redisCache, appLogger,
	)

	grpcServer := server.New(server.NewConfigMust(), appLogger)
	userv1.RegisterUserServiceServer(grpcServer.GRPC(), handler.NewHandler(usersService))

	if err := grpcServer.Run(ctx); err != nil {
		log.Fatalf("run grpc server: %v", err)
	}
}
