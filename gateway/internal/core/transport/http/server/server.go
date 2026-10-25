package core_http_server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	core_http_middleware "github.com/dyingvoid/shorturl/gateway/internal/core/transport/http/middleware"
	"github.com/go-chi/chi/v5"
)

const maxBodyBytes = 10 << 20

type Server struct {
	router     chi.Router
	cfg        Config
	log        *slog.Logger
	middleware []core_http_middleware.Middleware
}

func New(cfg Config, log *slog.Logger, middleware ...core_http_middleware.Middleware) *Server {
	return &Server{
		router:     chi.NewRouter(),
		cfg:        cfg,
		log:        log,
		middleware: middleware,
	}
}

func (s *Server) RegisterRoutes(routes ...Route) {
	registerRoutes(s.router, routes...)
}

func (s *Server) RegisterAPIRoutes(routers ...*VersionRouter) {
	for _, router := range routers {
		prefix := "/api/" + string(router.version)
		s.router.Mount(prefix, router.WithMiddleware())
	}
}

func (s *Server) Handler() http.Handler {
	handler := core_http_middleware.Chain(s.router, s.middleware...)
	return http.MaxBytesHandler(handler, maxBodyBytes)
}

func (s *Server) Run(ctx context.Context) error {
	handler := s.Handler()

	srv := &http.Server{
		Addr:              s.cfg.Address(),
		Handler:           handler,
		ReadHeaderTimeout: s.cfg.ReadHeaderTimeout,
		ReadTimeout:       s.cfg.ReadTimeout,
		WriteTimeout:      s.cfg.WriteTimeout,
		IdleTimeout:       s.cfg.IdleTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		s.log.Info("starting HTTP server", "addr", s.cfg.Address())
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("listen and serve HTTP: %w", err)
		}
		return nil
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			s.cfg.ShutdownTimeout,
		)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			_ = srv.Close()
			return fmt.Errorf("shutdown HTTP server: %w", err)
		}

		s.log.Info("HTTP server stopped")
		return nil
	}
}

func registerRoutes(router chi.Router, routes ...Route) {
	for _, route := range routes {
		handler := route.WithMiddleware()
		if route.Method == "" {
			router.Handle(route.Path, handler)
			continue
		}
		router.Method(route.Method, route.Path, handler)
	}
}
