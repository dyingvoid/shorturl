package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dyingvoid/shorturl/analytics/internal/config"
	"github.com/dyingvoid/shorturl/analytics/internal/core"
	"github.com/dyingvoid/shorturl/analytics/internal/handler"
	clickhouseinfra "github.com/dyingvoid/shorturl/analytics/internal/infrastructure/clickhouse"
	kafkainfra "github.com/dyingvoid/shorturl/analytics/internal/infrastructure/kafka"
	urlclient "github.com/dyingvoid/shorturl/analytics/internal/infrastructure/url_client"
	"github.com/dyingvoid/shorturl/analytics/internal/service"
	analyticsv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/analytics/v1"
	"github.com/dyingvoid/shorturl/shared/pkg/server"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := run(logger); err != nil {
		logger.Error("analytics stopped with error", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	ctx, stop := signal.NotifyContext(
		context.Background(), syscall.SIGINT, syscall.SIGTERM,
	)
	defer stop()

	cfg := config.NewMust()
	time.Local = cfg.TimeZone

	kafkaConfig := kafkainfra.NewMust()

	reader, err := kafkainfra.NewReader(kafkaConfig, logger)
	if err != nil {
		return fmt.Errorf("create kafka reader: %w", err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			logger.Error("close kafka reader", "error", err)
		}
	}()

	clickHouseConfig := clickhouseinfra.NewMust()

	clickHouseConn, err := clickhouseinfra.Connect(ctx, clickHouseConfig)
	if err != nil {
		return fmt.Errorf("connect clickhouse: %w", err)
	}
	defer func() {
		if err := clickHouseConn.Close(); err != nil {
			logger.Error("close clickhouse connection", "error", err)
		}
	}()

	analyticsRepository := clickhouseinfra.NewRepository(clickHouseConn)

	urlClient, err := urlclient.Connect(urlclient.NewMust())
	if err != nil {
		return fmt.Errorf("create url client: %w", err)
	}
	defer func() {
		if err := urlClient.Close(); err != nil {
			logger.Error("close url client", "error", err)
		}
	}()

	coreConfig := core.NewMust()

	consumer := core.NewAnalyticsForwarder(
		logger,
		reader,
		analyticsRepository,
		coreConfig.BatchSize,
		coreConfig.BatchFlushInterval,
	)

	analyticsService := service.New(urlClient, analyticsRepository)

	s := server.New(cfg, logger, "analytics")
	analyticsv1.RegisterAnalyticsServiceServer(s.GRPC(), handler.New(analyticsService))

	logger.Info(
		"analytics starting",
		"kafka_topic", kafkaConfig.Topic,
		"kafka_group", kafkaConfig.GroupID,
		"batch_size", coreConfig.BatchSize,
		"batch_flush_interval", coreConfig.BatchFlushInterval,
	)

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error, 2)

	go func() {
		errCh <- consumer.Run(runCtx)
	}()

	go func() {
		errCh <- s.Run(runCtx)
	}()

	err = <-errCh
	cancel()

	if err != nil {
		return fmt.Errorf("run analytics: %w", err)
	}

	return nil
}
