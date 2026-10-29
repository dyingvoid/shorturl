//go:build integration

package integration

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/require"
)

type tokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type clickEvent struct {
	ShortCode string    `json:"short_code"`
	UserID    string    `json:"user_id"`
	ClickedAt time.Time `json:"clicked_at"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Referer   string    `json:"referer"`
}

func TestShortURLFlowPublishesClickEvent(t *testing.T) {
	email := uniqueEmail(t)
	password := "Password1"
	originalURL := "https://example.com/articles/click-event"

	userID := suite.register(t, email, password)
	tokens := suite.login(t, email, password)
	shortCode := suite.createShortURL(t, tokens.AccessToken, originalURL)

	headers := map[string]string{
		"X-Forwarded-For": "203.0.113.42",
		"User-Agent":      "shorturl-integration-test/1.0",
		"Referer":         "https://referrer.example/source",
	}

	status, responseHeaders, _ := suite.request(
		t, http.MethodGet, "/"+shortCode, "", nil, headers,
	)
	require.Equal(t, http.StatusFound, status)
	require.Equal(t, originalURL, responseHeaders.Get("Location"))

	event := suite.waitForClickEvent(t, shortCode, 20*time.Second)
	require.Equal(t, shortCode, event.ShortCode)
	require.Equal(t, userID, event.UserID)
	require.Equal(t, "203.0.113.42", event.IP)
	require.Equal(t, "shorturl-integration-test/1.0", event.UserAgent)
	require.Equal(t, "https://referrer.example/source", event.Referer)
	require.False(t, event.ClickedAt.IsZero())
	require.WithinDuration(t, time.Now().UTC(), event.ClickedAt.UTC(), time.Minute)
}

func TestDeletedShortURLReturnsNotFound(t *testing.T) {
	email := uniqueEmail(t)
	password := "Password1"

	suite.register(t, email, password)
	tokens := suite.login(t, email, password)
	shortCode := suite.createShortURL(t, tokens.AccessToken, "https://example.com/articles/deleted")

	status, _, _ := suite.request(t, http.MethodGet, "/"+shortCode, "", nil, nil)
	require.Equal(t, http.StatusFound, status)

	status, _, body := suite.request(
		t, http.MethodDelete, "/api/v1/urls/"+shortCode, tokens.AccessToken, nil, nil,
	)
	require.Equal(t, http.StatusNoContent, status, string(body))

	status, _, _ = suite.request(t, http.MethodGet, "/"+shortCode, "", nil, nil)
	require.Equal(t, http.StatusNotFound, status)
}

func (s *testSuite) register(t *testing.T, email string, password string) string {
	t.Helper()

	status, _, body := s.request(
		t,
		http.MethodPost,
		"/api/v1/auth/register",
		"",
		map[string]string{"email": email, "password": password},
		nil,
	)
	require.Equal(t, http.StatusCreated, status, string(body))

	var response struct {
		UserID string `json:"user_id"`
	}
	require.NoError(t, json.Unmarshal(body, &response))
	require.NotEmpty(t, response.UserID)

	return response.UserID
}

func (s *testSuite) login(t *testing.T, email string, password string) tokenPair {
	t.Helper()

	status, _, body := s.request(
		t,
		http.MethodPost,
		"/api/v1/auth/login",
		"",
		map[string]string{"email": email, "password": password},
		nil,
	)
	require.Equal(t, http.StatusOK, status, string(body))

	var tokens tokenPair
	require.NoError(t, json.Unmarshal(body, &tokens))
	require.NotEmpty(t, tokens.AccessToken)
	require.NotEmpty(t, tokens.RefreshToken)

	return tokens
}

func (s *testSuite) createShortURL(t *testing.T, accessToken string, originalURL string) string {
	t.Helper()

	status, _, body := s.request(
		t,
		http.MethodPost,
		"/api/v1/urls",
		accessToken,
		map[string]string{"original_url": originalURL},
		nil,
	)
	require.Equal(t, http.StatusCreated, status, string(body))

	var response struct {
		ShortCode string `json:"short_code"`
		ShortURL  string `json:"short_url"`
	}
	require.NoError(t, json.Unmarshal(body, &response))
	require.NotEmpty(t, response.ShortCode)
	require.Contains(t, response.ShortURL, response.ShortCode)

	return response.ShortCode
}

func (s *testSuite) request(
	t *testing.T,
	method string,
	path string,
	accessToken string,
	body any,
	headers map[string]string,
) (int, http.Header, []byte) {
	t.Helper()

	var payload io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		require.NoError(t, err)
		payload = bytes.NewReader(raw)
	}

	req, err := http.NewRequest(method, s.baseURL+path, payload)
	require.NoError(t, err)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	response, err := s.client.Do(req)
	require.NoError(t, err)
	defer response.Body.Close()

	raw, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	return response.StatusCode, response.Header.Clone(), raw
}

func (s *testSuite) waitForClickEvent(t *testing.T, shortCode string, timeout time.Duration) clickEvent {
	t.Helper()

	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer(s.kafkaBrokers, config)
	require.NoError(t, err)
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition(kafkaTopic, 0, sarama.OffsetOldest)
	require.NoError(t, err)
	defer partitionConsumer.Close()

	deadline := time.NewTimer(timeout)
	defer deadline.Stop()

	for {
		select {
		case message, ok := <-partitionConsumer.Messages():
			if !ok {
				t.Fatal("kafka messages channel closed")
			}

			var event clickEvent
			require.NoError(t, json.Unmarshal(message.Value, &event))
			if event.ShortCode == shortCode {
				return event
			}
		case consumeErr, ok := <-partitionConsumer.Errors():
			if ok && consumeErr != nil {
				t.Fatalf("consume kafka message: %v", consumeErr)
			}
		case <-deadline.C:
			t.Fatalf("timed out waiting for click event %q", shortCode)
		}
	}
}

func uniqueEmail(t *testing.T) string {
	t.Helper()

	var random [8]byte
	_, err := rand.Read(random[:])
	require.NoError(t, err)

	return fmt.Sprintf("user-%s@example.com", hex.EncodeToString(random[:]))
}
