package mongo_errors

import (
	"errors"
	"fmt"

	domain_errors "github.com/dyingvoid/urlshort/url/internal/domain/errors"
	mongodriver "go.mongodb.org/mongo-driver/v2/mongo"
)

var ErrDuplicateKey = errors.New("duplicate key")

func MapError(err error) error {
	if err == nil {
		return nil
	}

	if mongodriver.IsDuplicateKeyError(err) {
		return fmt.Errorf(
			"%v: %w: %w",
			err,
			ErrDuplicateKey,
			domain_errors.ErrUniqueViolation,
		)
	}

	return err
}
