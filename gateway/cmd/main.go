package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/dyingvoid/shorturl/gateway/internal/config"
	core_http_middleware "github.com/dyingvoid/shorturl/gateway/internal/core/transport/http/middleware"
	core_http_server "github.com/dyingvoid/shorturl/gateway/internal/core/transport/http/server"
	auth_transport_http "github.com/dyingvoid/shorturl/gateway/internal/features/auth/transport/http"
	urls_transport_http "github.com/dyingvoid/shorturl/gateway/internal/features/urls/transport/http"
	"github.com/dyingvoid/shorturl/shared/pkg/logger"
	urlv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/url/v1"
	userv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/user/v1"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "github.com/dyingvoid/shorturl/gateway/docs"
)

// @title ShortURL Gateway API
// @version 1.0
// @description HTTP gateway for authentication and short URL management.
// @host localhost:8090
// @BasePath /api/v1
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Access token: "Bearer <token>".
func main() {
	// TODO: dynamic swagger registration (port is static rn, not from env)
	ctx, stop := signal.NotifyContext(
		context.Background(), syscall.SIGINT, syscall.SIGTERM,
	)
	defer stop()

	cfg := config.NewMust()
	httpConfig := core_http_server.MustNewConfig()
	log := logger.Init(logger.NewMust())
	log.Info("application starting")

	usersConn, err := grpc.NewClient(
		cfg.UsersServiceAddr(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Error("init users service client", "error", err)
		os.Exit(1)
	}
	defer usersConn.Close()

	urlsConn, err := grpc.NewClient(
		cfg.URLServiceAddr(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Error("init URL service client", "error", err)
		os.Exit(1)
	}
	defer urlsConn.Close()

	usersService := userv1.NewUserServiceClient(usersConn)
	urlsService := urlv1.NewURLServiceClient(urlsConn)

	authHTTP := auth_transport_http.NewAuthHTTPHandler(usersService)
	urlsHTTP := urls_transport_http.NewURLsHTTPHandler(urlsService, authHTTP.AuthMiddleware)
	urlRoutes := urlsHTTP.Routes()

	srv := core_http_server.New(httpConfig, log,
		core_http_middleware.RequestID(),
		core_http_middleware.Logger(log),
		core_http_middleware.Trace(),
		core_http_middleware.Recover(),
		core_http_middleware.CORS(httpConfig.AllowedOrigins),
		chimiddleware.Timeout(httpConfig.RequestTimeout),
	)

	v1 := core_http_server.NewVersionRouter(core_http_server.V1)
	v1.RegisterRoutes(authHTTP.Routes()...)
	v1.RegisterRoutes(urlRoutes[0]...)

	srv.RegisterAPIRoutes(v1)
	srv.RegisterRoutes(urlRoutes[1]...)
	srv.RegisterSwagger()

	if err := srv.Run(ctx); err != nil {
		log.Error("run HTTP server", "error", err)
		os.Exit(1)
	}
}
