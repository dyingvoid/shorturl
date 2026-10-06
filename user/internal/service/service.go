package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/dyingvoid/shorturl/user/internal/domain"
	"github.com/google/uuid"
)

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}

type TokenService interface {
	Issue(user domain.User, session domain.Session) (domain.TokenPair, error)
	ParseAccess(token string) (domain.TokenClaims, error)
	ParseRefresh(token string) (domain.TokenClaims, error)
	AccessTTL() time.Duration
	RefreshTTL() time.Duration
}

type UsersRepository interface {
	CreateUser(
		ctx context.Context,
		email domain.Email,
		passwordHash string,
	) (domain.User, error)
	GetUserByEmail(
		ctx context.Context,
		email domain.Email,
	) (domain.User, string, error)
}

type SessionsRepository interface {
	CreateSession(
		ctx context.Context,
		userID uuid.UUID,
		expiresAt time.Time,
	) (domain.Session, error)
	RevokeSession(
		ctx context.Context,
		id uuid.UUID,
	) error
}

type Cache interface {
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Get(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
}

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

func (s *UsersService) sessionKey(sessionID uuid.UUID) string {
	return fmt.Sprintf("session:%s", sessionID)
}
