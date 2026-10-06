package service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/dyingvoid/shorturl/user/internal/domain"
	domain_errors "github.com/dyingvoid/shorturl/user/internal/domain/errors"
	"github.com/dyingvoid/shorturl/user/internal/service"
	"github.com/dyingvoid/shorturl/user/internal/service/mocks"
)

func testTime() time.Time {
	return time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
}

func TestLogin_Success(t *testing.T) {
	t.Parallel()

	hasher := mocks.NewMockPasswordHasher(t)
	tokenService := mocks.NewMockTokenService(t)
	usersRepository := mocks.NewMockUsersRepository(t)
	sessionsRepository := mocks.NewMockSessionsRepository(t)
	cache := mocks.NewMockCache(t)

	credentials := domain.NewUserCredentials("user@example.com", "Password1")
	user := domain.NewUser(uuid.New(), credentials.Email.String(), testTime(), testTime())
	session := domain.NewSession(
		uuid.New(), user.ID, testTime().Add(time.Hour), testTime(), testTime(),
	)
	wantTokens := domain.TokenPair{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
	}

	usersRepository.EXPECT().
		GetUserByEmail(mock.Anything, credentials.Email).
		Return(user, "hashed-password", nil).
		Once()
	hasher.EXPECT().Compare("hashed-password", "Password1").Return(nil).Once()
	tokenService.EXPECT().RefreshTTL().Return(time.Hour).Once()
	sessionsRepository.EXPECT().
		CreateSession(mock.Anything, user.ID, mock.AnythingOfType("time.Time")).
		Return(session, nil).
		Once()
	tokenService.EXPECT().Issue(user, session).Return(wantTokens, nil).Once()
	tokenService.EXPECT().AccessTTL().Return(time.Minute).Once()
	cache.EXPECT().
		Set(mock.Anything, "session:"+session.ID.String(), user.ID.String(), time.Minute).
		Return(nil).
		Once()

	svc := service.NewUsersService(
		hasher, tokenService, usersRepository, sessionsRepository, cache, testLogger(),
	)

	got, err := svc.Login(context.Background(), credentials)

	require.NoError(t, err)
	assert.Equal(t, wantTokens, got)
}

func TestLogin_WrongPassword_Unauthenticated(t *testing.T) {
	t.Parallel()

	hasher := mocks.NewMockPasswordHasher(t)
	usersRepository := mocks.NewMockUsersRepository(t)

	credentials := domain.NewUserCredentials("user@example.com", "Password1")
	user := domain.NewUser(uuid.New(), credentials.Email.String(), testTime(), testTime())

	usersRepository.EXPECT().
		GetUserByEmail(mock.Anything, credentials.Email).
		Return(user, "hashed-password", nil).
		Once()
	hasher.EXPECT().
		Compare("hashed-password", "Password1").
		Return(errors.New("password mismatch")).
		Once()

	// tokenService/sessionsRepository/cache stay nil: the flow must stop after
	// the failed password comparison.
	svc := service.NewUsersService(
		hasher, nil, usersRepository, nil, nil, testLogger(),
	)

	got, err := svc.Login(context.Background(), credentials)

	require.Error(t, err)
	assert.ErrorIs(t, err, domain_errors.ErrUnauthenticated)
	assert.Equal(t, domain.TokenPair{}, got)
}

func TestLogin_UserNotFound_Unauthenticated(t *testing.T) {
	t.Parallel()

	hasher := mocks.NewMockPasswordHasher(t)
	usersRepository := mocks.NewMockUsersRepository(t)

	credentials := domain.NewUserCredentials("missing@example.com", "Password1")

	usersRepository.EXPECT().
		GetUserByEmail(mock.Anything, credentials.Email).
		Return(domain.User{}, "", fmt.Errorf("user query: %w", domain_errors.ErrNotFound)).
		Once()

	svc := service.NewUsersService(
		hasher, nil, usersRepository, nil, nil, testLogger(),
	)

	_, err := svc.Login(context.Background(), credentials)

	require.Error(t, err)
	assert.ErrorIs(t, err, domain_errors.ErrUnauthenticated)
}

func TestLogin_InvalidCredentials(t *testing.T) {
	t.Parallel()

	hasher := mocks.NewMockPasswordHasher(t)
	usersRepository := mocks.NewMockUsersRepository(t)

	credentials := domain.NewUserCredentials("user@example.com", "short")

	svc := service.NewUsersService(
		hasher, nil, usersRepository, nil, nil, testLogger(),
	)

	_, err := svc.Login(context.Background(), credentials)

	require.Error(t, err)
	assert.ErrorIs(t, err, domain_errors.ErrInvalidArgument)
}
