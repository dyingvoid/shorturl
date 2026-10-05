package service

import (
	"context"
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
}

type UsersService struct {
	passwordHasher     PasswordHasher
	tokenService       TokenService
	usersRepository    UsersRepository
	sessionsRepository SessionsRepository
}

func NewUsersService(
	passwordHasher PasswordHasher,
	tokenService TokenService,
	usersRepository UsersRepository,
	sessionsRepository SessionsRepository,
) *UsersService {
	return &UsersService{
		passwordHasher:     passwordHasher,
		tokenService:       tokenService,
		usersRepository:    usersRepository,
		sessionsRepository: sessionsRepository,
	}
}
