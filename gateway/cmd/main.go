package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/dyingvoid/shorturl/gateway/internal/config"
	core_http_middleware "github.com/dyingvoid/shorturl/gateway/internal/core/transport/http/middleware"
	core_http_server "github.com/dyingvoid/shorturl/gateway/internal/core/transport/http/server"
	analytics_transport_http "github.com/dyingvoid/shorturl/gateway/internal/features/analytics/transport/http"
	auth_transport_http "github.com/dyingvoid/shorturl/gateway/internal/features/auth/transport/http"
	urls_transport_http "github.com/dyingvoid/shorturl/gateway/internal/features/urls/transport/http"
	"github.com/dyingvoid/shorturl/shared/pkg/logger"
	analyticsv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/analytics/v1"
	urlv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/url/v1"
	userv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/user/v1"
	"github.com/dyingvoid/shorturl/shared/pkg/requestid"
	"github.com/dyingvoid/shorturl/shared/pkg/tracing"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	_ "github.com/dyingvoid/shorturl/gateway/docs"
)

// @title ShortURL Gateway API
// @version 1.0
// @description HTTP gateway for authentication and short URL management.
// @BasePath /
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Access token: "Bearer <token>".
func main() {
	log := logger.Init(logger.NewMust())
	log.Warn("application started")

	if err := run(log); err != nil {
		log.Error("gateway stopped with error", "error", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	// TODO: dynamic swagger registration (port is static rn, not from env)
	ctx, stop := signal.NotifyContext(
		context.Background(), syscall.SIGINT, syscall.SIGTERM,
	)
	defer stop()

	cfg := config.NewMust()
	httpConfig := core_http_server.MustNewConfig()

	traceProvider, err := tracing.New(ctx, log, "gateway")
	if err != nil {
		return fmt.Errorf("init tracing: %w", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			httpConfig.ShutdownTimeout,
		)
		defer cancel()

		if err := traceProvider.Shutdown(shutdownCtx); err != nil {
			log.Error("shutdown tracing", "error", err)
		}
	}()

	usersConn, err := grpc.NewClient(
		cfg.UsersServiceAddr(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(requestid.UnaryClientInterceptor()),
		grpc.WithStatsHandler(tracing.GRPCClientHandler()),
	)
	if err != nil {
		return fmt.Errorf("init users service client: %w", err)
	}
	defer func() {
		if err := usersConn.Close(); err != nil {
			log.Error("close users service client", "error", err)
		}
	}()

	urlsConn, err := grpc.NewClient(
		cfg.URLServiceAddr(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(requestid.UnaryClientInterceptor()),
		grpc.WithStatsHandler(tracing.GRPCClientHandler()),
	)
	if err != nil {
		return fmt.Errorf("init URL service client: %w", err)
	}
	defer func() {
		if err := urlsConn.Close(); err != nil {
			log.Error("close URL service client", "error", err)
		}
	}()

	analyticsConn, err := grpc.NewClient(
		cfg.AnalyticsServiceAddr(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(tracing.GRPCClientHandler()),
	)
	if err != nil {
		return fmt.Errorf("init analytics service client: %w", err)
	}
	defer func() {
		if err := analyticsConn.Close(); err != nil {
			log.Error("close analytics service client", "error", err)
		}
	}()

	usersService := userv1.NewUserServiceClient(usersConn)
	urlsService := urlv1.NewURLServiceClient(urlsConn)
	analyticsService := analyticsv1.NewAnalyticsServiceClient(analyticsConn)

	authHTTP := auth_transport_http.NewAuthHTTPHandler(usersService)
	urlsHTTP := urls_transport_http.NewURLsHTTPHandler(urlsService, authHTTP.AuthMiddleware)
	analyticsHTTP := analytics_transport_http.NewAnalyticsHTTPHandler(analyticsService, authHTTP.AuthMiddleware)
	urlRoutes := urlsHTTP.Routes()

	registry := prometheus.NewRegistry()
	srv, err := core_http_server.New(
		httpConfig, log, registry,
		core_http_middleware.Tracing(),
		core_http_middleware.RequestID(),
		core_http_middleware.Logger(log),
		core_http_middleware.Trace(),
		core_http_middleware.Recover(),
		core_http_middleware.CORS(httpConfig.AllowedOrigins),
		chimiddleware.Timeout(httpConfig.RequestTimeout),
	)
	if err != nil {
		return fmt.Errorf("create HTTP server: %w", err)
	}

	v1 := core_http_server.NewVersionRouter(core_http_server.V1)
	v1.RegisterRoutes(authHTTP.Routes()...)
	v1.RegisterRoutes(urlRoutes[0]...)
	v1.RegisterRoutes(analyticsHTTP.Routes()...)

	srv.RegisterAPIRoutes(v1)
	srv.RegisterRoutes(urlRoutes[1]...)
	srv.RegisterSwagger()

	return srv.Run(ctx)
}
