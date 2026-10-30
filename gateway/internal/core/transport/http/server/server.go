package core_http_server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	core_http_middleware "github.com/dyingvoid/shorturl/gateway/internal/core/transport/http/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const maxBodyBytes = 10 << 20

type Server struct {
	router        chi.Router
	metricsServer *http.Server
	cfg           Config
	log           *slog.Logger
	middleware    []core_http_middleware.Middleware
}

func New(
	cfg Config,
	log *slog.Logger,
	reg MetricsRegistry,
	middleware ...core_http_middleware.Middleware,
) (*Server, error) {
	s := &Server{
		router:     chi.NewRouter(),
		cfg:        cfg,
		log:        log,
		middleware: middleware,
	}

	if reg != nil {
		m, err := newMetrics(reg)
		if err != nil {
			return nil, fmt.Errorf("create metrics: %w", err)
		}

		s.middleware = append(s.middleware, m.Middleware())

		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.HandlerFor(
			reg,
			promhttp.HandlerOpts{},
		))
		s.metricsServer = &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.MetricsPort),
			Handler: mux,
		}
	}

	return s, nil
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

	httpErrCh := make(chan error, 1)
	metricsErrCh := make(chan error, 1)

	go func() {
		s.log.Info("starting HTTP server", "addr", s.cfg.Address())
		if err := srv.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			httpErrCh <- fmt.Errorf("listen and serve HTTP: %w", err)
		}
	}()

	if s.metricsServer != nil {
		go func() {
			s.log.Info(
				"starting metrics server",
				"addr", s.metricsServer.Addr,
			)
			if err := s.metricsServer.ListenAndServe(); err != nil &&
				!errors.Is(err, http.ErrServerClosed) {
				metricsErrCh <- fmt.Errorf(
					"listen and serve metrics: %w",
					err,
				)
			}
		}()
	}

	select {
	case err := <-httpErrCh:
		return errors.Join(err, s.stop(srv))

	case err := <-metricsErrCh:
		return errors.Join(err, s.stop(srv))

	case <-ctx.Done():
		return s.stop(srv)
	}
}

func (s *Server) stop(srv *http.Server) error {
	s.log.Info("shutdown HTTP server")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		s.cfg.ShutdownTimeout,
	)
	defer cancel()

	httpErr := s.httpStop(shutdownCtx, srv)
	metricsErr := s.metricsStop(shutdownCtx)

	s.log.Info("HTTP server stopped")

	return errors.Join(httpErr, metricsErr)
}

func (s *Server) httpStop(ctx context.Context, srv *http.Server) error {
	if err := srv.Shutdown(ctx); err != nil {
		_ = srv.Close()
		return fmt.Errorf("shutdown HTTP server: %w", err)
	}

	return nil
}

func (s *Server) metricsStop(ctx context.Context) error {
	if s.metricsServer == nil {
		return nil
	}

	if err := s.metricsServer.Shutdown(ctx); err != nil {
		_ = s.metricsServer.Close()
		return fmt.Errorf("shutdown metrics server: %w", err)
	}

	return nil
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
