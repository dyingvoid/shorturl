package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func (s *UsersService) GetLimit(
	ctx context.Context,
	userID uuid.UUID,
) (int, error) {
	limit, err := s.usersRepository.GetLimit(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("get limit: %w", err)
	}

	return limit, nil
}
