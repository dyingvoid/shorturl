package mongo_repository_outbox

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	outboxCollection = "outbox"
	eventURLClicked  = "url.clicked"
)

type OutboxRepository struct {
	collection *mongo.Collection
}

func NewOutboxRepository(db *mongo.Database) *OutboxRepository {
	return &OutboxRepository{
		collection: db.Collection(outboxCollection),
	}
}
func (r *OutboxRepository) PutuserStatistics(
	ctx context.Context,
	ip, userAgent, referer string,
) error {
	model := newUserStatistics(ip, userAgent, referer)
	if _, err := r.collection.InsertOne(ctx, model); err != nil {
		return err
	}

	return nil
}

func (r *OutboxRepository) GetUnpublished(
	ctx context.Context,
	limit int64,
) ([]outboxModel[userStatistics], error) {
	if limit <= 0 {
		return nil, errors.New("limit must be positive")
	}

	opts := options.Find().
		SetSort(bson.D{
			{Key: "created_at", Value: 1},
			{Key: "_id", Value: 1},
		}).
		SetLimit(limit)

	cursor, err := r.collection.Find(ctx, bson.D{
		{Key: "published_at", Value: nil},
	}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []outboxModel[userStatistics]
	if err := cursor.All(ctx, &events); err != nil {
		return nil, err
	}

	return events, nil
}
