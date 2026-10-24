package links_mongo_repository

import (
	"context"
	"fmt"

	"github.com/dyingvoid/urlshort/url/internal/domain"
	domain_errors "github.com/dyingvoid/urlshort/url/internal/domain/errors"
	mongo_errors "github.com/dyingvoid/urlshort/url/internal/infrastructure/mongo/errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	linksCollection = "links"
)

type LinksRepository struct {
	db *mongo.Database
}

func NewLinksRepository(db *mongo.Database) *LinksRepository {
	return &LinksRepository{db: db}
}

func (r *LinksRepository) Count(ctx context.Context, userID string) (int, error) {
	filter := bson.M{"user_id": userID}

	count, err := r.db.Collection(linksCollection).CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("link count: %w", mongo_errors.MapError(err))
	}

	return int(count), nil
}

func (r *LinksRepository) Insert(ctx context.Context, link *domain.Link) error {
	collection := r.db.Collection(linksCollection)

	_, err := collection.InsertOne(ctx, newLinkModel(*link))
	if err != nil {
		return fmt.Errorf("link insert: %w", mongo_errors.MapError(err))
	}

	return nil
}

func (r *LinksRepository) GetByShortCode(ctx context.Context, shortCode string) (*domain.Link, error) {
	filter := bson.M{"short_code": shortCode}

	var model linkModel
	if err := r.db.Collection(linksCollection).FindOne(ctx, filter).Decode(&model); err != nil {
		return nil, fmt.Errorf("link find: %w", mongo_errors.MapError(err))
	}

	link := model.ToDomain()

	return &link, nil
}

func (r *LinksRepository) ListByUser(
	ctx context.Context,
	userID string,
	limit int,
	cursor string,
) ([]domain.Link, error) {
	filter := bson.M{"user_id": userID}

	if cursor != "" {
		objectID, err := bson.ObjectIDFromHex(cursor)
		if err != nil {
			return nil, fmt.Errorf(
				"invalid cursor '%s': %v: %w",
				cursor,
				err,
				domain_errors.ErrInvalidArgument,
			)
		}

		filter["_id"] = bson.M{"$gt": objectID}
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "_id", Value: 1}}).
		SetLimit(int64(limit + 1))

	findCursor, err := r.db.Collection(linksCollection).Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("link find: %w", mongo_errors.MapError(err))
	}
	defer findCursor.Close(ctx)

	var models []linkModel
	if err := findCursor.All(ctx, &models); err != nil {
		return nil, fmt.Errorf("link collect: %w", mongo_errors.MapError(err))
	}

	links := make([]domain.Link, 0, len(models))
	for _, model := range models {
		links = append(links, model.ToDomain())
	}

	return links, nil
}

func (r *LinksRepository) Delete(ctx context.Context, shortCode string) error {
	filter := bson.M{"short_code": shortCode}

	result, err := r.db.Collection(linksCollection).DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("link delete: %w", mongo_errors.MapError(err))
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("link delete: %w", domain_errors.ErrNotFound)
	}

	return nil
}
