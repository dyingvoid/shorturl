package service

import (
	"context"
	"fmt"

	"github.com/dyingvoid/shorturl/user/internal/domain"
	domain_errors "github.com/dyingvoid/shorturl/user/internal/domain/errors"
)

func (s *UsersService) Login(
	ctx context.Context,
	credentials domain.UserCredentials,
) (domain.TokenPair, error) {
	if err := credentials.Validate(); err != nil {
		return domain.TokenPair{}, fmt.Errorf("invalid credentials: %w", err)
	}

	user, passwordHash, err := s.usersRepository.GetUserByEmail(ctx, credentials.Email)
	if err != nil {
		return domain.TokenPair{}, fmt.Errorf(
			"get user by email: %v: %w", err, domain_errors.ErrUnauthenticated,
		)
	}

	if err := s.passwordHasher.Compare(passwordHash, credentials.Password.String()); err != nil {
		return domain.TokenPair{}, fmt.Errorf(
			"compare password: %v: %w", err, domain_errors.ErrUnauthenticated,
		)
	}

	tokens, err := s.tokenService.Issue(user)
	if err != nil {
		return domain.TokenPair{}, fmt.Errorf("issue tokens: %w", err)
	}

	return tokens, nil
}
