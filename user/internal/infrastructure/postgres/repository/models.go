package users_postgres_repository

import (
	"time"

	"github.com/dyingvoid/shorturl/user/internal/domain"
	"github.com/google/uuid"
)

type userModel struct {
	ID           uuid.UUID `db:"id"`
	Email        string    `db:"email"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

func (m userModel) ToDomain() domain.User {
	return domain.NewUser(m.ID, m.Email, m.CreatedAt, m.UpdatedAt)
}

type sessionModel struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	ExpiresAt time.Time `db:"expires_at"`
	IsRevoked bool      `db:"is_revoked"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (m sessionModel) ToDomain() domain.Session {
	return domain.NewSession(m.ID, m.UserID, m.ExpiresAt, m.CreatedAt, m.UpdatedAt)
}
