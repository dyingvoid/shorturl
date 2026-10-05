package main

import (
	"context"
	"fmt"
	"log"
	"net"

	userv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/user/v1"
	"github.com/dyingvoid/shorturl/user/internal/config"
	"github.com/dyingvoid/shorturl/user/internal/handler"
	"github.com/dyingvoid/shorturl/user/internal/infrastructure/hashing"
	"github.com/dyingvoid/shorturl/user/internal/infrastructure/jwt"
	"github.com/dyingvoid/shorturl/user/internal/infrastructure/logger"
	"github.com/dyingvoid/shorturl/user/internal/infrastructure/postgres"
	users_postgres_repository "github.com/dyingvoid/shorturl/user/internal/infrastructure/postgres/repository"
	"github.com/dyingvoid/shorturl/user/internal/infrastructure/redis"
	"github.com/dyingvoid/shorturl/user/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	ctx := context.Background()

	appLogger := logger.Init(logger.NewMust())
	appLogger.Info("application started")

	pool, err := postgres.Connect(ctx, postgres.NewMust())
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer pool.Close()

	redisCache, err := redis.Connect(ctx, redis.NewMust())
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
		redisCache,
	)

	serverCfg := config.NewMust()
	grpcServer := grpc.NewServer()
	userv1.RegisterUserServiceServer(grpcServer, handler.NewHandler(usersService))

	reflection.Register(grpcServer)
	appLogger.Warn("server starting", "port", serverCfg.UserServicePort)
	lis, err := net.Listen(
		"tcp", fmt.Sprintf(":%d", serverCfg.UserServicePort),
	)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	log.Fatal(grpcServer.Serve(lis))
}
