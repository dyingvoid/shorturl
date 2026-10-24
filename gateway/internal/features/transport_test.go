package features_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	core_http_middleware "github.com/dyingvoid/shorturl/gateway/internal/core/transport/http/middleware"
	core_http_server "github.com/dyingvoid/shorturl/gateway/internal/core/transport/http/server"
	auth_transport_http "github.com/dyingvoid/shorturl/gateway/internal/features/auth/transport/http"
	urls_transport_http "github.com/dyingvoid/shorturl/gateway/internal/features/urls/transport/http"
	urlv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/url/v1"
	userv1 "github.com/dyingvoid/shorturl/shared/pkg/proto/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeUsers struct {
	userv1.UserServiceClient
	register        func(*userv1.RegisterRequest) (*userv1.RegisterResponse, error)
	validateSession func(*userv1.ValidateSessionRequest) (*userv1.ValidateSessionResponse, error)
}

func (f *fakeUsers) Register(
	_ context.Context,
	req *userv1.RegisterRequest,
	_ ...grpc.CallOption,
) (*userv1.RegisterResponse, error) {
	return f.register(req)
}

func (f *fakeUsers) ValidateSession(
	_ context.Context,
	req *userv1.ValidateSessionRequest,
	_ ...grpc.CallOption,
) (*userv1.ValidateSessionResponse, error) {
	return f.validateSession(req)
}

type fakeURLs struct {
	urlv1.URLServiceClient
	redirect       func(*urlv1.RedirectRequest) (*urlv1.RedirectResponse, error)
	createShortURL func(*urlv1.CreateShortURLRequest) (*urlv1.CreateShortURLResponse, error)
}

func (f *fakeURLs) Redirect(
	_ context.Context,
	req *urlv1.RedirectRequest,
	_ ...grpc.CallOption,
) (*urlv1.RedirectResponse, error) {
	return f.redirect(req)
}

func (f *fakeURLs) CreateShortURL(
	_ context.Context,
	req *urlv1.CreateShortURLRequest,
	_ ...grpc.CallOption,
) (*urlv1.CreateShortURLResponse, error) {
	return f.createShortURL(req)
}

func TestRoutes(t *testing.T) {
	users := &fakeUsers{}
	urls := &fakeURLs{}

	srv := newTestServer(t, users, urls)
	client := &http.Client{
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	t.Run("register", func(t *testing.T) {
		users.register = func(req *userv1.RegisterRequest) (*userv1.RegisterResponse, error) {
			if req.GetEmail() != "user@example.com" || req.GetPassword() != "Secret1" {
				t.Errorf("unexpected credentials: %v", req)
			}
			return &userv1.RegisterResponse{UserId: "user-1"}, nil
		}

		resp, err := client.Post(
			srv.URL+"/api/v1/auth/register",
			"application/json",
			strings.NewReader(`{"email":"user@example.com","password":"Secret1"}`),
		)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusCreated)
		}

		var body map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["user_id"] != "user-1" {
			t.Fatalf("user_id = %q, want %q", body["user_id"], "user-1")
		}
	})

	t.Run("redirect", func(t *testing.T) {
		urls.redirect = func(req *urlv1.RedirectRequest) (*urlv1.RedirectResponse, error) {
			if req.GetShortCode() != "abc123" {
				t.Errorf("short code = %q, want %q", req.GetShortCode(), "abc123")
			}
			return &urlv1.RedirectResponse{OriginalUrl: "https://example.com/target"}, nil
		}

		resp, err := client.Get(srv.URL + "/abc123")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusFound {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusFound)
		}
		if location := resp.Header.Get("Location"); location != "https://example.com/target" {
			t.Fatalf("location = %q, want %q", location, "https://example.com/target")
		}
	})

	t.Run("redirect not found", func(t *testing.T) {
		urls.redirect = func(*urlv1.RedirectRequest) (*urlv1.RedirectResponse, error) {
			return nil, status.Error(codes.NotFound, "link not found")
		}

		resp, err := client.Get(srv.URL + "/missing")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
		}

		var body map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["error"] != "link not found" {
			t.Fatalf("error = %q, want %q", body["error"], "link not found")
		}
	})

	t.Run("protected route without token", func(t *testing.T) {
		resp, err := client.Post(
			srv.URL+"/api/v1/urls",
			"application/json",
			strings.NewReader(`{"original_url":"https://example.com"}`),
		)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
		}
	})

	t.Run("protected route with token", func(t *testing.T) {
		users.validateSession = func(req *userv1.ValidateSessionRequest) (*userv1.ValidateSessionResponse, error) {
			if req.GetAccessToken() != "access-token" {
				t.Errorf("access token = %q, want %q", req.GetAccessToken(), "access-token")
			}
			return &userv1.ValidateSessionResponse{
				UserId:    "user-1",
				SessionId: "session-1",
			}, nil
		}
		urls.createShortURL = func(req *urlv1.CreateShortURLRequest) (*urlv1.CreateShortURLResponse, error) {
			if req.GetUserId() != "user-1" {
				t.Errorf("user id = %q, want %q", req.GetUserId(), "user-1")
			}
			if req.GetOriginalUrl() != "https://example.com" {
				t.Errorf("original URL = %q, want %q", req.GetOriginalUrl(), "https://example.com")
			}
			return &urlv1.CreateShortURLResponse{
				ShortCode: "abc123",
				ShortUrl:  "http://localhost/abc123",
			}, nil
		}

		req, err := http.NewRequest(
			http.MethodPost,
			srv.URL+"/api/v1/urls",
			strings.NewReader(`{"original_url":"https://example.com"}`),
		)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer access-token")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusCreated)
		}

		var body map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["short_code"] != "abc123" {
			t.Fatalf("short_code = %q, want %q", body["short_code"], "abc123")
		}
	})
}

func newTestServer(t *testing.T, users *fakeUsers, urls *fakeURLs) *httptest.Server {
	t.Helper()

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	authHTTP := auth_transport_http.NewAuthHTTPHandler(users)
	urlsHTTP := urls_transport_http.NewURLsHTTPHandler(urls, authHTTP.AuthMiddleware)

	srv := core_http_server.New(core_http_server.Config{Addr: "127.0.0.1:0"}, log,
		core_http_middleware.RequestID(),
		core_http_middleware.Logger(log),
		core_http_middleware.Trace(),
		core_http_middleware.Recover(),
	)

	v1 := core_http_server.NewVersionRouter(core_http_server.V1)
	v1.RegisterRoutes(authHTTP.Routes()...)
	v1.RegisterRoutes(urlsHTTP.Routes()...)
	srv.RegisterAPIRoutes(v1)
	srv.RegisterRoutes(urlsHTTP.PublicRoutes()...)

	testServer := httptest.NewServer(srv.Handler())
	t.Cleanup(testServer.Close)

	return testServer
}
