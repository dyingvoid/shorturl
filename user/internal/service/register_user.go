package service

import (
	"context"
	"fmt"

	"github.com/dyingvoid/shorturl/user/internal/domain"
)

func (s *UsersService) Register(
	ctx context.Context,
	credentials domain.UserCredentials,
) (domain.User, error) {
	if err := credentials.Validate(); err != nil {
		return domain.User{}, fmt.Errorf("invalid credentials: %w", err)
	}

	passHash, err := s.passwordHasher.Hash(credentials.Password.String())
	if err != nil {
		return domain.User{}, fmt.Errorf("invalid credentials: %w", err)
	}

	user, err := s.usersRepository.CreateUser(
		ctx, credentials.Email, passHash,
	)
	if err != nil {
		return domain.User{}, fmt.Errorf("failed to add user: %w", err)
	}

	return user, nil
}
