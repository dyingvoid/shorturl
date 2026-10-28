package kafka

import (
	"errors"
	"strings"
)

type Config struct {
	Brokers []string
	Topic   string
	GroupID string
}

func NewConfig(brokers []string, topic, groupID string) (Config, error) {
	cleaned := make([]string, 0, len(brokers))
	for _, broker := range brokers {
		if broker = strings.TrimSpace(broker); broker != "" {
			cleaned = append(cleaned, broker)
		}
	}

	cfg := Config{
		Brokers: cleaned,
		Topic:   strings.TrimSpace(topic),
		GroupID: strings.TrimSpace(groupID),
	}

	if len(cfg.Brokers) == 0 {
		return Config{}, errors.New("kafka: at least one broker is required")
	}
	if cfg.Topic == "" {
		return Config{}, errors.New("kafka: topic is required")
	}
	if cfg.GroupID == "" {
		return Config{}, errors.New("kafka: consumer group ID is required")
	}

	return cfg, nil
}
