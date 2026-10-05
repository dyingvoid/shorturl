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
