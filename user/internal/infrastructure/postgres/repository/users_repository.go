package users_postgres_repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/dyingvoid/shorturl/user/internal/domain"
	domain_errors "github.com/dyingvoid/shorturl/user/internal/domain/errors"
	postgres "github.com/dyingvoid/shorturl/user/internal/infrastructure/postgres/pool"
)

type UsersRepository struct {
	pool postgres.Pool
}

func NewUsersRepository(pool postgres.Pool) *UsersRepository {
	return &UsersRepository{pool: pool}
}

func (r *UsersRepository) CreateUser(
	ctx context.Context,
	email domain.Email,
	passwordHash string,
) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	WITH new_user AS (
		INSERT INTO users.users (email, password_hash)
		VALUES (@email, @password_hash)
		RETURNING *
	), new_subscription AS (
		INSERT INTO users.subscriptions (user_id, plan, links_limit)
		SELECT id, 'basic', 1000 FROM new_user
	)
	SELECT * FROM new_user;`

	args := pgx.NamedArgs{
		"email":         email.String(),
		"password_hash": passwordHash,
	}

	rows, err := r.pool.Query(ctx, query, args)
	if err != nil {
		return domain.User{}, fmt.Errorf("user query: %w", err)
	}

	userModel, err := postgres.CollectExactlyOneRow[userModel](rows)
	if err != nil {
		return domain.User{}, fmt.Errorf("user collect: %w", err)
	}

	return userModel.ToDomain(), nil
}

func (r *UsersRepository) GetUserByEmail(
	ctx context.Context,
	email domain.Email,
) (domain.User, string, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	SELECT *
	FROM users.users
	WHERE email = @email;`

	args := pgx.NamedArgs{
		"email": email.String(),
	}

	rows, err := r.pool.Query(ctx, query, args)
	if err != nil {
		return domain.User{}, "", fmt.Errorf("user query: %w", err)
	}

	userModel, err := postgres.CollectExactlyOneRow[userModel](rows)
	if err != nil {
		if errors.Is(err, domain_errors.ErrNotFound) {
			return domain.User{}, "", fmt.Errorf(
				"user with email='%s' not found: %w",
				email.String(),
				domain_errors.ErrNotFound,
			)
		}
		return domain.User{}, "", fmt.Errorf("user collect: %w", err)
	}

	return userModel.ToDomain(), userModel.PasswordHash, nil
}

func (r *UsersRepository) GetLimit(
	ctx context.Context,
	userID uuid.UUID,
) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	query := `
	SELECT links_limit
	FROM users.subscriptions
	WHERE user_id = @user_id;`

	args := pgx.NamedArgs{
		"user_id": userID,
	}

	rows, err := r.pool.Query(ctx, query, args)
	if err != nil {
		return 0, fmt.Errorf("subscription query: %w", err)
	}

	subscriptionModel, err := postgres.CollectExactlyOneRow[subscriptionModel](rows)
	if err != nil {
		if errors.Is(err, domain_errors.ErrNotFound) {
			return 0, fmt.Errorf(
				"subscription with user_id='%s' not found: %w",
				userID,
				domain_errors.ErrNotFound,
			)
		}
		return 0, fmt.Errorf("subscription collect: %w", err)
	}

	return subscriptionModel.LinksLimit, nil
}
