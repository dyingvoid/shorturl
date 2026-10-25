package mongo_repository_outbox

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (r *OutboxRepository) EnsureIndexes(ctx context.Context) error {
	model := mongo.IndexModel{
		Keys: bson.D{
			{Key: "created_at", Value: 1},
			{Key: "_id", Value: 1},
		},
		Options: options.Index().
			SetName("idx_outbox_unpublished").
			SetPartialFilterExpression(bson.D{
				{Key: "published_at", Value: nil},
			}),
	}

	if _, err := r.collection.Indexes().CreateOne(ctx, model); err != nil {
		return err
	}

	return nil
}
