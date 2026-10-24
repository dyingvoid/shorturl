package service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	domain_errors "github.com/dyingvoid/urlshort/url/internal/domain/errors"
	"github.com/dyingvoid/urlshort/url/internal/service"
	"github.com/dyingvoid/urlshort/url/internal/service/mocks"
)

const (
	testBaseURL       = "https://short.local"
	testUserID        = "user-1"
	testOriginalURL   = "https://example.com/some/very/long/path"
	testCreateRetries = 3
)

func newTestService(
	t *testing.T,
	usersService service.UsersService,
	cache service.Cache,
	links service.LinksRepository,
) *service.Service {
	t.Helper()

	return service.New(
		testBaseURL,
		testCreateRetries,
		time.Minute,
		usersService,
		cache,
		links,
	)
}

func TestCreateShortURL_Success(t *testing.T) {
	t.Parallel()

	usersService := mocks.NewMockUsersService(t)
	cache := mocks.NewMockCache(t)
	links := mocks.NewMockLinksRepository(t)

	usersService.EXPECT().GetLimit(mock.Anything, testUserID).Return(10, nil).Once()
	cache.EXPECT().GetLinkCounter(mock.Anything, testUserID).Return(4, nil).Once()
	cache.EXPECT().IncLinkCounter(mock.Anything, testUserID).Return(5, nil).Once()
	links.EXPECT().
		Insert(mock.Anything, mock.AnythingOfType("*domain.Link")).
		Return(nil).
		Once()

	svc := newTestService(t, usersService, cache, links)

	got, err := svc.CreateShortURL(context.Background(), testUserID, testOriginalURL, nil)

	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, testUserID, got.UserID)
	assert.Equal(t, testOriginalURL, got.OriginalURL)
	assert.NotEmpty(t, got.ShortCode)
	assert.True(t, got.Active())

	cache.AssertNotCalled(t, "DecLinkCounter", mock.Anything, mock.Anything)
}

func TestCreateShortURL_LimitExhausted_RollsBackCounter(t *testing.T) {
	t.Parallel()

	usersService := mocks.NewMockUsersService(t)
	cache := mocks.NewMockCache(t)
	links := mocks.NewMockLinksRepository(t)

	const limit = 2

	usersService.EXPECT().GetLimit(mock.Anything, testUserID).Return(limit, nil).Once()
	cache.EXPECT().GetLinkCounter(mock.Anything, testUserID).Return(limit, nil).Once()
	cache.EXPECT().IncLinkCounter(mock.Anything, testUserID).Return(limit+1, nil).Once()
	cache.EXPECT().DecLinkCounter(mock.Anything, testUserID).Return(limit, nil).Once()

	svc := newTestService(t, usersService, cache, links)

	got, err := svc.CreateShortURL(context.Background(), testUserID, testOriginalURL, nil)

	require.Error(t, err)
	assert.ErrorIs(t, err, domain_errors.ErrResourceExhausted)
	assert.Nil(t, got)

	links.AssertNotCalled(t, "Insert", mock.Anything, mock.Anything)
}

func TestCreateShortURL_InsertFailure_RollsBackCounter(t *testing.T) {
	t.Parallel()

	usersService := mocks.NewMockUsersService(t)
	cache := mocks.NewMockCache(t)
	links := mocks.NewMockLinksRepository(t)

	insertErr := errors.New("mongo is down")

	usersService.EXPECT().GetLimit(mock.Anything, testUserID).Return(10, nil).Once()
	cache.EXPECT().GetLinkCounter(mock.Anything, testUserID).Return(4, nil).Once()
	cache.EXPECT().IncLinkCounter(mock.Anything, testUserID).Return(5, nil).Once()
	links.EXPECT().
		Insert(mock.Anything, mock.AnythingOfType("*domain.Link")).
		Return(insertErr).
		Once()
	cache.EXPECT().DecLinkCounter(mock.Anything, testUserID).Return(4, nil).Once()

	svc := newTestService(t, usersService, cache, links)

	got, err := svc.CreateShortURL(context.Background(), testUserID, testOriginalURL, nil)

	require.Error(t, err)
	assert.ErrorIs(t, err, insertErr)
	assert.Nil(t, got)
}

func TestCreateShortURL_UniqueViolation_RetriesWithoutRollback(t *testing.T) {
	t.Parallel()

	usersService := mocks.NewMockUsersService(t)
	cache := mocks.NewMockCache(t)
	links := mocks.NewMockLinksRepository(t)

	uniqueErr := fmt.Errorf("link insert: %w", domain_errors.ErrUniqueViolation)

	usersService.EXPECT().GetLimit(mock.Anything, testUserID).Return(10, nil).Once()
	cache.EXPECT().GetLinkCounter(mock.Anything, testUserID).Return(4, nil).Once()
	cache.EXPECT().IncLinkCounter(mock.Anything, testUserID).Return(5, nil).Once()
	links.EXPECT().
		Insert(mock.Anything, mock.AnythingOfType("*domain.Link")).
		Return(uniqueErr).
		Once()
	links.EXPECT().
		Insert(mock.Anything, mock.AnythingOfType("*domain.Link")).
		Return(nil).
		Once()

	svc := newTestService(t, usersService, cache, links)

	got, err := svc.CreateShortURL(context.Background(), testUserID, testOriginalURL, nil)

	require.NoError(t, err)
	require.NotNil(t, got)

	// A retried collision still ends in a commit: the counter stays incremented.
	cache.AssertNotCalled(t, "DecLinkCounter", mock.Anything, mock.Anything)
}

func TestCreateShortURL_CounterCacheMiss_InitializesFromRepository(t *testing.T) {
	t.Parallel()

	usersService := mocks.NewMockUsersService(t)
	cache := mocks.NewMockCache(t)
	links := mocks.NewMockLinksRepository(t)

	usersService.EXPECT().GetLimit(mock.Anything, testUserID).Return(10, nil).Once()
	cache.EXPECT().
		GetLinkCounter(mock.Anything, testUserID).
		Return(0, fmt.Errorf("cache miss: %w", domain_errors.ErrNotFound)).
		Once()
	links.EXPECT().Count(mock.Anything, testUserID).Return(3, nil).Once()
	cache.EXPECT().SetLinkCounter(mock.Anything, testUserID, 3).Return(nil).Once()
	cache.EXPECT().IncLinkCounter(mock.Anything, testUserID).Return(4, nil).Once()
	links.EXPECT().
		Insert(mock.Anything, mock.AnythingOfType("*domain.Link")).
		Return(nil).
		Once()

	svc := newTestService(t, usersService, cache, links)

	got, err := svc.CreateShortURL(context.Background(), testUserID, testOriginalURL, nil)

	require.NoError(t, err)
	require.NotNil(t, got)
}
