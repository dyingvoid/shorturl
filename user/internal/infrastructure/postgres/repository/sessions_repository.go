package users_postgres_repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dyingvoid/shorturl/user/internal/domain"
	"github.com/dyingvoid/shorturl/user/internal/infrastructure/postgres"
)

type SessionsRepository struct {
	pool postgres.Pool
}

func NewSessionsRepository(pool postgres.Pool) *SessionsRepository {
	return &SessionsRepository{pool: pool}
}

func (r *SessionsRepository) CreateSession(
	ctx context.Context,
	userID uuid.UUID,
	expiresAt time.Time,
) (domain.Session, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	INSERT INTO users.sessions (user_id, expires_at)
	VALUES (@user_id, @expires_at)
	RETURNING *;`

	args := pgx.NamedArgs{
		"user_id":    userID,
		"expires_at": expiresAt,
	}

	rows, err := r.pool.Query(ctx, query, args)
	if err != nil {
		return domain.Session{}, fmt.Errorf("session query: %w", err)
	}

	sessionModel, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[sessionModel])
	if err != nil {
		return domain.Session{}, fmt.Errorf("session collect: %w", err)
	}

	return sessionModel.ToDomain(), nil
}
