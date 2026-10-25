package mongo_repository_outbox

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type outboxModel[T any] struct {
	ObjectID    bson.ObjectID `bson:"_id,omitempty"`
	EventType   string        `bson:"event_type"`
	Payload     T             `bson:"payload"`
	CreatedAt   time.Time     `bson:"created_at"`
	PublishedAt *time.Time    `bson:"published_at"`
}

type userStatistics struct {
	IP        string `bson:"ip"`
	UserAgent string `bson:"user_agent"`
	Referer   string `bson:"referer"`
}

func newUserStatistics(ip, userAgent, referer string) outboxModel[userStatistics] {
	return outboxModel[userStatistics]{
		EventType: eventURLClicked,
		Payload: userStatistics{
			IP:        ip,
			UserAgent: userAgent,
			Referer:   referer,
		},
		CreatedAt: time.Now().UTC(),
	}
}
