package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

func Connect(ctx context.Context, cfg Config) (*mongo.Database, error) {
	client, err := mongo.Connect(options.Client().
		ApplyURI(uri(cfg)).
		SetAuth(options.Credential{
			Username:   cfg.User,
			Password:   cfg.Password,
			AuthSource: cfg.AuthSource,
		}).
		SetConnectTimeout(cfg.Timeout).
		SetServerSelectionTimeout(cfg.Timeout))
	if err != nil {
		return nil, fmt.Errorf("connect mongo: %w", err)
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		_ = client.Disconnect(ctx)

		return nil, fmt.Errorf("ping mongo: %w", err)
	}

	return client.Database(cfg.DB), nil
}

func uri(cfg Config) string {
	return fmt.Sprintf("mongodb://%s:%d", cfg.Host, cfg.Port)
}
