package service

import (
	"fmt"
	"log/slog"

	"github.com/google/uuid"
)

type UsersService struct {
	passwordHasher     PasswordHasher
	tokenService       TokenService
	usersRepository    UsersRepository
	sessionsRepository SessionsRepository
	cache              Cache
	logger             *slog.Logger
}

func NewUsersService(
	passwordHasher PasswordHasher,
	tokenService TokenService,
	usersRepository UsersRepository,
	sessionsRepository SessionsRepository,
	cache Cache,
	logger *slog.Logger,
) *UsersService {
	return &UsersService{
		passwordHasher:     passwordHasher,
		tokenService:       tokenService,
		usersRepository:    usersRepository,
		sessionsRepository: sessionsRepository,
		cache:              cache,
		logger:             logger,
	}
}

const sessionKeyPrefix = "session:"

func (s *UsersService) sessionKey(sessionID uuid.UUID) string {
	return fmt.Sprintf("%s%s", sessionKeyPrefix, sessionID)
}
