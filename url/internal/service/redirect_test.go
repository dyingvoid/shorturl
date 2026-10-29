package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/dyingvoid/urlshort/url/internal/domain"
	"github.com/dyingvoid/urlshort/url/internal/service/mocks"
)

const testShortCode = "abc1234"

func TestRedirect_ProducesUserView(t *testing.T) {
	t.Parallel()

	cache := mocks.NewMockCache(t)
	links := mocks.NewMockLinksRepository(t)
	producer := mocks.NewMockProducer(t)

	userView := domain.NewUserView("127.0.0.1", "test-agent", "https://referer.local")
	link := &domain.Link{
		UserID:      "owner-1",
		OriginalURL: testOriginalURL,
		ShortCode:   testShortCode,
		IsActive:    true,
	}

	cache.EXPECT().GetLink(mock.Anything, testShortCode).Return(link, nil).Once()
	cache.EXPECT().SetLink(mock.Anything, link, time.Minute).Return(nil).Once()

	produced := make(chan domain.UserView, 1)
	producer.EXPECT().ProduceUserView(mock.Anything).Run(func(view domain.UserView) {
		produced <- view
	}).Return(nil).Once()

	svc := newTestServiceWithProducer(t, nil, cache, links, producer)

	got, err := svc.Redirect(context.Background(), testShortCode, userView)

	require.NoError(t, err)
	assert.Equal(t, testOriginalURL, got.Value)

	select {
	case view := <-produced:
		assert.Equal(t, testShortCode, view.ShortCode)
		assert.Equal(t, "owner-1", view.UserID)
		assert.False(t, view.ClickedAt.IsZero())
		assert.Equal(t, userView.IP, view.IP)
		assert.Equal(t, userView.UserAgent, view.UserAgent)
		assert.Equal(t, userView.Referer, view.Referer)
	case <-time.After(time.Second):
		t.Fatal("user view was not produced")
	}
}

func TestRedirect_ProducerErrorIsIgnored(t *testing.T) {
	t.Parallel()

	cache := mocks.NewMockCache(t)
	links := mocks.NewMockLinksRepository(t)
	producer := mocks.NewMockProducer(t)

	userView := domain.NewUserView("127.0.0.1", "test-agent", "https://referer.local")
	link := &domain.Link{
		UserID:      "owner-1",
		OriginalURL: testOriginalURL,
		ShortCode:   testShortCode,
		IsActive:    true,
	}

	cache.EXPECT().GetLink(mock.Anything, testShortCode).Return(link, nil).Once()
	cache.EXPECT().SetLink(mock.Anything, link, time.Minute).Return(nil).Once()

	produced := make(chan struct{}, 1)
	producer.EXPECT().ProduceUserView(mock.Anything).Run(func(domain.UserView) {
		produced <- struct{}{}
	}).Return(errors.New("kafka is down")).Once()

	svc := newTestServiceWithProducer(t, nil, cache, links, producer)

	got, err := svc.Redirect(context.Background(), testShortCode, userView)

	require.NoError(t, err)
	assert.Equal(t, testOriginalURL, got.Value)

	select {
	case <-produced:
	case <-time.After(time.Second):
		t.Fatal("user view was not produced")
	}
}

func TestRedirect_InactiveLinkStillProducesUserView(t *testing.T) {
	t.Parallel()

	cache := mocks.NewMockCache(t)
	links := mocks.NewMockLinksRepository(t)
	producer := mocks.NewMockProducer(t)

	userView := domain.NewUserView("127.0.0.1", "test-agent", "https://referer.local")
	link := &domain.Link{
		UserID:      "owner-1",
		OriginalURL: testOriginalURL,
		ShortCode:   testShortCode,
		IsActive:    false,
	}

	cache.EXPECT().GetLink(mock.Anything, testShortCode).Return(link, nil).Once()

	produced := make(chan domain.UserView, 1)
	producer.EXPECT().ProduceUserView(mock.Anything).Run(func(view domain.UserView) {
		produced <- view
	}).Return(nil).Once()

	svc := newTestServiceWithProducer(t, nil, cache, links, producer)

	got, err := svc.Redirect(context.Background(), testShortCode, userView)

	require.Error(t, err)
	assert.Empty(t, got.Value)

	select {
	case view := <-produced:
		assert.Equal(t, testShortCode, view.ShortCode)
		assert.Equal(t, "owner-1", view.UserID)
		assert.False(t, view.ClickedAt.IsZero())
	case <-time.After(time.Second):
		t.Fatal("user view was not produced")
	}
}
