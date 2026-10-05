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
	"github.com/dyingvoid/shorturl/user/internal/infrastructure/postgres"
	users_postgres_repository "github.com/dyingvoid/shorturl/user/internal/infrastructure/postgres/repository"
	"github.com/dyingvoid/shorturl/user/internal/service"
	"google.golang.org/grpc"
)

func main() {
	ctx := context.Background()

	pool, err := postgres.Connect(ctx, postgres.NewMust())
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer pool.Close()

	hasher := hashing.NewPasswordHasher()
	usersRepository := users_postgres_repository.NewUsersRepository(pool)
	usersService := service.NewUsersService(hasher, usersRepository)

	grpcServer := grpc.NewServer()
	userv1.RegisterUserServiceServer(grpcServer, handler.NewHandler(usersService))

	lis, err := net.Listen(
		"tcp", fmt.Sprintf(":%d", config.NewMust().UserServicePort),
	)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	log.Fatal(grpcServer.Serve(lis))
}
