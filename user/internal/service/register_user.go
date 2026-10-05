package service

import (
	"context"
	"fmt"

	"github.com/dyingvoid/shorturl/user/internal/domain"
)

func (s *UsersService) RegisterUser(
	ctx context.Context,
	credentials domain.UserCredentials,
) (domain.User, error) {
	if err := credentials.Validate(); err != nil {
		return domain.User{}, fmt.Errorf("invalid credentials: %w", err)
	}

	return domain.User{}, nil
}
