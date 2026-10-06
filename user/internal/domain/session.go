package domain

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	IsRevoked bool
	ExpiresAt time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewSession(
	id, userID uuid.UUID,
	expiresAt, createdAt, updatedAt time.Time,
) Session {
	return Session{
		ID:        id,
		UserID:    userID,
		ExpiresAt: expiresAt,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

func (s *Session) IsActive() bool {
	return !s.IsRevoked && s.ExpiresAt.After(time.Now())
}
