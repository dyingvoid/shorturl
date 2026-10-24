package service_test

import (
	"context"
	"fmt"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/dyingvoid/shorturl/user/internal/domain"
	domain_errors "github.com/dyingvoid/shorturl/user/internal/domain/errors"
	"github.com/dyingvoid/shorturl/user/internal/service"
	"github.com/dyingvoid/shorturl/user/internal/service/mocks"
)

func testLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

func TestRegister_Success(t *testing.T) {
	t.Parallel()

	hasher := mocks.NewMockPasswordHasher(t)
	usersRepository := mocks.NewMockUsersRepository(t)

	credentials := domain.NewUserCredentials("user@example.com", "Password1")
	wantUser := domain.NewUser(uuid.New(), credentials.Email.String(), testTime(), testTime())

	hasher.EXPECT().Hash("Password1").Return("hashed-password", nil).Once()
	usersRepository.EXPECT().
		CreateUser(mock.Anything, credentials.Email, "hashed-password").
		Return(wantUser, nil).
		Once()

	svc := service.NewUsersService(
		hasher, nil, usersRepository, nil, nil, testLogger(),
	)

	got, err := svc.Register(context.Background(), credentials)

	require.NoError(t, err)
	assert.Equal(t, wantUser, got)
}

func TestRegister_DuplicateEmail_AlreadyExists(t *testing.T) {
	t.Parallel()

	hasher := mocks.NewMockPasswordHasher(t)
	usersRepository := mocks.NewMockUsersRepository(t)

	credentials := domain.NewUserCredentials("user@example.com", "Password1")
	repositoryErr := fmt.Errorf("unique violation: %w", domain_errors.ErrAlreadyExists)

	hasher.EXPECT().Hash("Password1").Return("hashed-password", nil).Once()
	usersRepository.EXPECT().
		CreateUser(mock.Anything, credentials.Email, "hashed-password").
		Return(domain.User{}, repositoryErr).
		Once()

	svc := service.NewUsersService(
		hasher, nil, usersRepository, nil, nil, testLogger(),
	)

	got, err := svc.Register(context.Background(), credentials)

	require.Error(t, err)
	assert.ErrorIs(t, err, domain_errors.ErrAlreadyExists)
	assert.Equal(t, domain.User{}, got)
}

func TestRegister_InvalidCredentials(t *testing.T) {
	t.Parallel()

	hasher := mocks.NewMockPasswordHasher(t)
	usersRepository := mocks.NewMockUsersRepository(t)

	// Password fails validation, so hashing and persistence must not happen.
	credentials := domain.NewUserCredentials("user@example.com", "short")

	svc := service.NewUsersService(
		hasher, nil, usersRepository, nil, nil, testLogger(),
	)

	_, err := svc.Register(context.Background(), credentials)

	require.Error(t, err)
	assert.ErrorIs(t, err, domain_errors.ErrInvalidArgument)
}

func TestRegister_HashFailure(t *testing.T) {
	t.Parallel()

	hasher := mocks.NewMockPasswordHasher(t)
	usersRepository := mocks.NewMockUsersRepository(t)

	credentials := domain.NewUserCredentials("user@example.com", "Password1")

	hasher.EXPECT().Hash("Password1").Return("", fmt.Errorf("hash boom")).Once()

	svc := service.NewUsersService(
		hasher, nil, usersRepository, nil, nil, testLogger(),
	)

	_, err := svc.Register(context.Background(), credentials)

	require.Error(t, err)
	assert.ErrorContains(t, err, "hash boom")
}
