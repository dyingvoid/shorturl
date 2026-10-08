package links_mongo_repository

import (
	"time"

	"github.com/dyingvoid/urlshort/url/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type linkModel struct {
	ObjectID    bson.ObjectID `bson:"_id,omitempty"`
	UserID      string        `bson:"user_id"`
	ShortCode   string        `bson:"short_code"`
	OriginalURL string        `bson:"original_url"`
	ExpiresAt   *time.Time    `bson:"expires_at,omitempty"`
	CreatedAt   time.Time     `bson:"created_at"`
	IsActive    bool          `bson:"is_active"`
}

func (m linkModel) ToDomain() domain.Link {
	return domain.Link{
		ID:          m.ObjectID.Hex(),
		UserID:      m.UserID,
		OriginalURL: m.OriginalURL,
		ExpiresAt:   m.ExpiresAt,
	}
}

func newLinkModel(link domain.Link) linkModel {
	var oid bson.ObjectID

	if link.ID == "" {
		oid = bson.NewObjectID()
	} else {
		var err error
		oid, err = bson.ObjectIDFromHex(link.ID)
		if err != nil {
			panic(err)
		}
	}

	return linkModel{
		ObjectID:    oid,
		UserID:      link.UserID,
		ShortCode:   link.ShortCode,
		OriginalURL: link.OriginalURL,
		ExpiresAt:   link.ExpiresAt,
		CreatedAt:   time.Now(),
		IsActive:    true,
	}
}
