package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/dyingvoid/shorturl/analytics/internal/application"
	"github.com/dyingvoid/shorturl/analytics/internal/config"
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

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	kafkaConfig, err := kafkainfra.NewConfig(
		cfg.KafkaBrokers, cfg.KafkaTopic, cfg.KafkaGroupID,
	)
	if err != nil {
		return fmt.Errorf("create kafka config: %w", err)
	}

	reader, err := kafkainfra.NewReader(kafkaConfig)
	if err != nil {
		return fmt.Errorf("create kafka reader: %w", err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			logger.Error("close kafka reader", "error", err)
		}
	}()

	clickHouseConfig, err := clickhouseinfra.NewConfig(
		cfg.ClickHouseAddr,
		cfg.ClickHouseDatabase,
		cfg.ClickHouseUsername,
		cfg.ClickHousePassword,
	)
	if err != nil {
		return fmt.Errorf("create clickhouse config: %w", err)
	}

	writer, err := clickhouseinfra.NewWriter(ctx, clickHouseConfig)
	if err != nil {
		return fmt.Errorf("create clickhouse writer: %w", err)
	}
	defer func() {
		if err := writer.Close(); err != nil {
			logger.Error("close clickhouse writer", "error", err)
		}
	}()

	consumer := application.NewConsumer(
		reader,
		writer,
		cfg.BatchSize,
		cfg.BatchFlushInterval,
		logger,
	)

	logger.Info(
		"analytics consumer starting",
		"kafka_topic", cfg.KafkaTopic,
		"kafka_group", cfg.KafkaGroupID,
		"batch_size", cfg.BatchSize,
		"batch_flush_interval", cfg.BatchFlushInterval,
	)

	if err := consumer.Run(ctx); err != nil {
		return fmt.Errorf("run consumer: %w", err)
	}

	return nil
}
