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
	clickhouseinfra "github.com/dyingvoid/shorturl/analytics/internal/infrastructure/clickhouse"
	kafkainfra "github.com/dyingvoid/shorturl/analytics/internal/infrastructure/kafka"
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

	writer, err := clickhouseinfra.NewWriter(ctx, clickHouseConfig)
	if err != nil {
		return fmt.Errorf("create clickhouse writer: %w", err)
	}
	defer func() {
		if err := writer.Close(); err != nil {
			logger.Error("close clickhouse writer", "error", err)
		}
	}()

	coreConfig := core.NewMust()

	consumer := core.NewAnalyticsForwarder(
		logger,
		reader,
		writer,
		coreConfig.BatchSize,
		coreConfig.BatchFlushInterval,
	)

	logger.Info(
		"analytics consumer starting",
		"kafka_topic", kafkaConfig.Topic,
		"kafka_group", kafkaConfig.GroupID,
		"batch_size", coreConfig.BatchSize,
		"batch_flush_interval", coreConfig.BatchFlushInterval,
	)

	if err := consumer.Run(ctx); err != nil {
		return fmt.Errorf("run consumer: %w", err)
	}

	return nil
}
