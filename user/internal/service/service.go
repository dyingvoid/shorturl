package service

import (
	"context"

	"github.com/dyingvoid/shorturl/user/internal/domain"
)

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}

type TokenService interface {
	Issue(user domain.User) (domain.TokenPair, error)
	ParseAccess(token string) (domain.TokenClaims, error)
	ParseRefresh(token string) (domain.TokenClaims, error)
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

type UsersService struct {
	passwordHasher  PasswordHasher
	tokenService    TokenService
	usersRepository UsersRepository
}

func NewUsersService(
	passwordHasher PasswordHasher,
	tokenService TokenService,
	usersRepository UsersRepository,
) *UsersService {
	return &UsersService{
		passwordHasher:  passwordHasher,
		tokenService:    tokenService,
		usersRepository: usersRepository,
	}
}
