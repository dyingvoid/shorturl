package service

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, password string) error
}

type UsersService struct {
	passwordHasher PasswordHasher
}

func NewUsersService(passwordHasher PasswordHasher) *UsersService {
	return &UsersService{
		passwordHasher: passwordHasher,
	}
}
