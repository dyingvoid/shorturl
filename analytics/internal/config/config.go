package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const envFile = ".env"

type Config struct {
	KafkaBrokers []string
	KafkaTopic   string
	KafkaGroupID string

	ClickHouseAddr     string
	ClickHouseDatabase string
	ClickHouseUsername string
	ClickHousePassword string

	BatchSize          int
	BatchFlushInterval time.Duration
}

func Load() (Config, error) {
	if err := loadEnvFile(envFile); err != nil {
		return Config{}, err
	}

	batchSize, err := strconv.Atoi(getEnv("BATCH_SIZE", "100"))
	if err != nil {
		return Config{}, fmt.Errorf("BATCH_SIZE must be an integer: %w", err)
	}

	flushInterval, err := time.ParseDuration(getEnv("BATCH_FLUSH_INTERVAL", "5s"))
	if err != nil {
		return Config{}, fmt.Errorf("BATCH_FLUSH_INTERVAL must be a duration: %w", err)
	}

	cfg := Config{
		KafkaBrokers:       splitList(getEnv("KAFKA_BROKERS", "localhost:9092")),
		KafkaTopic:         getEnv("KAFKA_TOPIC", "click_events"),
		KafkaGroupID:       getEnv("KAFKA_GROUP_ID", "analytics"),
		ClickHouseAddr:     getEnv("CLICKHOUSE_ADDR", "localhost:9000"),
		ClickHouseDatabase: getEnv("CLICKHOUSE_DATABASE", "analytics"),
		ClickHouseUsername: getEnv("CLICKHOUSE_USERNAME", "default"),
		ClickHousePassword: getEnv("CLICKHOUSE_PASSWORD", ""),
		BatchSize:          batchSize,
		BatchFlushInterval: flushInterval,
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) validate() error {
	if len(c.KafkaBrokers) == 0 {
		return errors.New("config: KAFKA_BROKERS must contain at least one broker")
	}
	if c.KafkaTopic == "" {
		return errors.New("config: KAFKA_TOPIC is required")
	}
	if c.KafkaGroupID == "" {
		return errors.New("config: KAFKA_GROUP_ID is required")
	}
	if c.ClickHouseAddr == "" {
		return errors.New("config: CLICKHOUSE_ADDR is required")
	}
	if c.ClickHouseDatabase == "" {
		return errors.New("config: CLICKHOUSE_DATABASE is required")
	}
	if c.ClickHouseUsername == "" {
		return errors.New("config: CLICKHOUSE_USERNAME is required")
	}
	if c.BatchSize <= 0 {
		return errors.New("config: BATCH_SIZE must be greater than zero")
	}
	if c.BatchFlushInterval <= 0 {
		return errors.New("config: BATCH_FLUSH_INTERVAL must be greater than zero")
	}

	return nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func splitList(value string) []string {
	parts := strings.Split(value, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			values = append(values, part)
		}
	}
	return values
}

func loadEnvFile(path string) error {
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("open env file %q: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		line = strings.TrimPrefix(line, "export ")
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return fmt.Errorf("env file %q line %d: expected KEY=VALUE", path, lineNumber)
		}

		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key == "" {
			return fmt.Errorf("env file %q line %d: empty key", path, lineNumber)
		}
		if _, exists := os.LookupEnv(key); exists {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("set env %q: %w", key, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read env file %q: %w", path, err)
	}

	return nil
}
