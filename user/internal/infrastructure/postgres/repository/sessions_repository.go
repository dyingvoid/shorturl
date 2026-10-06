package users_postgres_repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dyingvoid/shorturl/user/internal/domain"
	domain_errors "github.com/dyingvoid/shorturl/user/internal/domain/errors"
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

	sessionModel, err := postgres.CollectExactlyOneRow[sessionModel](rows)
	if err != nil {
		return domain.Session{}, fmt.Errorf("session collect: %w", err)
	}

	return sessionModel.ToDomain(), nil
}

func (r *SessionsRepository) GetSession(
	ctx context.Context,
	id uuid.UUID,
) (domain.Session, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	SELECT *
	FROM users.sessions
	WHERE id = @id;`

	args := pgx.NamedArgs{
		"id": id,
	}

	rows, err := r.pool.Query(ctx, query, args)
	if err != nil {
		return domain.Session{}, fmt.Errorf("session query: %w", err)
	}

	sessionModel, err := postgres.CollectExactlyOneRow[sessionModel](rows)
	if err != nil {
		if errors.Is(err, domain_errors.ErrNotFound) {
			return domain.Session{}, fmt.Errorf(
				"active session with id='%s' not found: %w",
				id,
				domain_errors.ErrNotFound,
			)
		}
		return domain.Session{}, fmt.Errorf("session collect: %w", err)
	}

	return sessionModel.ToDomain(), nil
}

func (r *SessionsRepository) RevokeSession(
	ctx context.Context,
	id uuid.UUID,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	UPDATE users.sessions
	SET is_revoked = TRUE, updated_at = NOW()
	WHERE id = @id;`

	args := pgx.NamedArgs{
		"id": id,
	}

	tag, err := r.pool.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("session update: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf(
			"session with id='%s' not found: %w",
			id,
			domain_errors.ErrNotFound,
		)
	}

	return nil
}
