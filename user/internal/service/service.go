package service

import (
	"context"

	"github.com/dyingvoid/shorturl/user/internal/domain"
)

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}

type UsersRepository interface {
	CreateUser(
		ctx context.Context,
		email domain.Email,
		passwordHash string,
	) (domain.User, error)
}

type UsersService struct {
	passwordHasher  PasswordHasher
	usersRepository UsersRepository
}

func NewUsersService(
	passwordHasher PasswordHasher,
	usersRepository UsersRepository,
) *UsersService {
	return &UsersService{
		passwordHasher:  passwordHasher,
		usersRepository: usersRepository,
	}
}
