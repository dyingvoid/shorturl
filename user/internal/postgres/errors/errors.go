package postgres_errors

import "errors"

const PgForeignKeyViolation = "23503"

var ErrFKViolation = errors.New("foreign key violation")
