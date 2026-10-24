package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewUser(
	id uuid.UUID,
	email string,
	createdAt, updatedAt time.Time,
) User {
	return User{
		ID:        id,
		Email:     email,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}
