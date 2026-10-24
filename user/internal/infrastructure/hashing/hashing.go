package hashing

import (
	"fmt"

	domain_errors "github.com/dyingvoid/shorturl/user/internal/domain/errors"
	"golang.org/x/crypto/bcrypt"
)

const cost = 10

type PasswordHasher struct{}

func NewPasswordHasher() *PasswordHasher {
	return &PasswordHasher{}
}

func (h *PasswordHasher) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("generate password hash: %w", err)
	}

	return string(hash), nil
}

func (h *PasswordHasher) Compare(hash, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return fmt.Errorf("compare password hash: %v: %w", err, domain_errors.ErrInvalidArgument)
	}

	return nil
}
