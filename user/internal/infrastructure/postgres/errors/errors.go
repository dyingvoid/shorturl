package postgres_errors

import (
	"errors"
	"fmt"

	domain_errors "github.com/dyingvoid/shorturl/user/internal/domain/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const PgForeignKeyViolation = "23503"
const PgUniqueViolation = "23505"

var (
	ErrFKViolation     = errors.New("foreign key violation")
	ErrUniqueViolation = errors.New("unique violation")
)

func MapError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%w: %w", err, domain_errors.ErrNotFound)
	}

	var outErr error
	var domainErr error
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case PgForeignKeyViolation:
			outErr = ErrFKViolation
			domainErr = domain_errors.ErrNotFound
		case PgUniqueViolation:
			outErr = ErrUniqueViolation
			domainErr = domain_errors.ErrAlreadyExists
		}
	}

	return fmt.Errorf(
		"%v: %w: %w",
		err,
		outErr,
		domainErr,
	)
}
