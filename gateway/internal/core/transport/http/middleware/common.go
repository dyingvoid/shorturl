package core_http_middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	core_http_response "github.com/dyingvoid/shorturl/gateway/internal/core/transport/http/response"
	"github.com/dyingvoid/shorturl/shared/pkg/logger"
	"github.com/dyingvoid/shorturl/shared/pkg/requestid"
)

func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get(requestid.Header)
			if id == "" {
				id = newRequestID()
			}

			r.Header.Set(requestid.Header, id)
			w.Header().Set(requestid.Header, id)

			ctx := requestid.ToContext(r.Context(), id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func Logger(base *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := base.With(
				slog.String("request_id", r.Header.Get(requestid.Header)),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
			)

			ctx := logger.ToContext(r.Context(), log)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func Trace() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := logger.FromContext(r.Context())
			rw := core_http_response.NewWriter(w)

			start := time.Now()
			log.Debug("request started", "method", r.Method, "path", r.URL.Path)
			next.ServeHTTP(rw, r)
			log.Info(
				"request finished",
				"status", rw.Status(),
				"duration", time.Since(start),
			)
		})
	}
}

func Recover() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := logger.FromContext(r.Context())
			response := core_http_response.NewHandler(log, w)

			defer func() {
				if p := recover(); p != nil {
					response.Panic(p, "unexpected panic during HTTP request")
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

func CORS(allowedOrigins []string) Middleware {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowed[origin] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" {
				w.Header().Add("Vary", "Origin")
			}

			if _, ok := allowed[origin]; ok && origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
				w.Header().Set("Access-Control-Max-Age", "300")
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}
	return hex.EncodeToString(b[:])
}
