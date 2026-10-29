package main

import (
	"context"
	"errors"
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
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	cfg := config.NewMust()
	time.Local = cfg.TimeZone

	kafkaConfig := kafkainfra.NewMust()
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

	urlClient, err := urlclient.Connect(urlclient.NewMust())
	if err != nil {
		return fmt.Errorf("create url client: %w", err)
	}
	defer func() {
		if err := urlClient.Close(); err != nil {
			logger.Error("close url client", "error", err)
		}
	}()

	reader, err := kafkainfra.NewReader(kafkaConfig, logger)
	if err != nil {
		return fmt.Errorf("create kafka reader: %w", err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			logger.Error("close kafka reader", "error", err)
		}
	}()

	analyticsRepository := clickhouseinfra.NewRepository(clickHouseConn)
	coreConfig := core.NewMust()

	consumer := core.NewAnalyticsForwarder(
		logger,
		reader,
		analyticsRepository,
		coreConfig.BatchSize,
		coreConfig.BatchFlushInterval,
	)

	analyticsService := core.NewService(
		coreConfig,
		urlClient,
		analyticsRepository,
	)

	s := server.New(cfg, logger, "analytics")
	analyticsv1.RegisterAnalyticsServiceServer(
		s.GRPC(),
		handler.New(analyticsService),
	)

	logger.Info(
		"analytics starting",
		"kafka_topic", kafkaConfig.Topic,
		"kafka_group", kafkaConfig.GroupID,
		"batch_size", coreConfig.BatchSize,
		"batch_flush_interval", coreConfig.BatchFlushInterval,
		"default_days", coreConfig.DefaultDays,
		"max_days", coreConfig.MaxDays,
		"top_referers", coreConfig.TopReferers,
	)

	serverCtx, stopServer := context.WithCancel(context.Background())
	defer stopServer()

	consumerCtx, stopConsumer := context.WithCancel(context.Background())
	defer stopConsumer()

	serverErrCh := make(chan error, 1)
	consumerErrCh := make(chan error, 1)

	go func() {
		serverErrCh <- s.Run(serverCtx)
	}()

	go func() {
		consumerErrCh <- consumer.Run(consumerCtx)
	}()

	var (
		runErr       error
		serverDone   bool
		consumerDone bool
	)

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")

	case err := <-serverErrCh:
		serverDone = true
		runErr = err

	case err := <-consumerErrCh:
		consumerDone = true
		runErr = err
	}

	logger.Info("stopping analytics server")
	stopServer()

	if !serverDone {
		if err := <-serverErrCh; err != nil &&
			!errors.Is(err, context.Canceled) &&
			runErr == nil {
			runErr = err
		}
	}

	logger.Info("stopping analytics consumer")
	stopConsumer()

	if !consumerDone {
		if err := <-consumerErrCh; err != nil &&
			!errors.Is(err, context.Canceled) &&
			runErr == nil {
			runErr = err
		}
	}

	if errors.Is(runErr, context.Canceled) {
		runErr = nil
	}

	if runErr != nil {
		return fmt.Errorf("run analytics: %w", runErr)
	}

	logger.Info("analytics stopped gracefully")
	return nil
}
