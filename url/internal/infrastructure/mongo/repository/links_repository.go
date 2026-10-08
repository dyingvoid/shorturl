package links_mongo_repository

import (
	"context"
	"fmt"

	"github.com/dyingvoid/urlshort/url/internal/domain"
	mongo_errors "github.com/dyingvoid/urlshort/url/internal/infrastructure/mongo/errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
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
