package users_postgres_repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/dyingvoid/shorturl/user/internal/domain"
	"github.com/dyingvoid/shorturl/user/internal/infrastructure/postgres"
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

	userModel, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[userModel])
	if err != nil {
		return domain.User{}, fmt.Errorf("user collect: %w", err)
	}

	return userModel.ToDomain(), nil
}
