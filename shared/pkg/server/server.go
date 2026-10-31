package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/dyingvoid/shorturl/shared/pkg/tracing"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	grpc          *grpc.Server
	metricsServer *http.Server

	cfg     Config
	tracing *tracing.Provider

	log *slog.Logger
	id  string
}

type Config interface {
	Port() int
	MetricsPort() int
	ShutdownTimeout() time.Duration
}

func New(
	ctx context.Context,
	cfg Config,
	log *slog.Logger,
	id string,
	reg MetricsRegistry,
) (*Server, error) {
	log = log.With("server_id", id)

	tracer, err := tracing.New(ctx, log, id)
	if err != nil {
		return nil, fmt.Errorf("init tracing: %w", err)
	}

	server := &Server{
		cfg:     cfg,
		tracing: tracer,
		log:     log,
		id:      id,
	}
	interceptors := []grpc.UnaryServerInterceptor{
		RequestID(),
		Logger(log),
		Trace(),
		Recovery(log),
	}

	if reg != nil {
		m, err := newMetrics(reg)
		if err != nil {
			_ = tracer.Shutdown(ctx)
			return nil, fmt.Errorf("create metrics: %w", err)
		}

		interceptors = append(interceptors, m.UnaryServerInterceptor())
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.HandlerFor(
			reg,
			promhttp.HandlerOpts{},
		))
		server.metricsServer = &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.MetricsPort()),
			Handler: mux,
		}
	}

	server.grpc = grpc.NewServer(
		grpc.ChainUnaryInterceptor(interceptors...),
		grpc.StatsHandler(tracing.GRPCServerHandler()),
	)
	reflection.Register(server.grpc)

	return server, nil
}

func (s *Server) GRPC() *grpc.Server {
	return s.grpc
}

func (s *Server) Run(ctx context.Context) error {
	s.log.Warn(
		"server starting",
		"grpc port", s.cfg.Port(),
		"metrics port", s.cfg.MetricsPort(),
	)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.cfg.Port()))
	if err != nil {
		return fmt.Errorf("listen grpc: %w", err)
	}

	grpcErrCh := make(chan error, 1)
	metricsErrCh := make(chan error, 1)

	go func() {
		if s.metricsServer != nil {
			defer close(metricsErrCh)
			if err := s.metricsServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
				metricsErrCh <- err
			}
		}
	}()
	go func() {
		defer close(grpcErrCh)
		if err := s.grpc.Serve(lis); err != nil &&
			!errors.Is(err, grpc.ErrServerStopped) {
			grpcErrCh <- err
		}
	}()

	select {
	case grpcErr := <-grpcErrCh:
		return errors.Join(grpcErr, s.stop())

	case metricsErr := <-metricsErrCh:
		return errors.Join(metricsErr, s.stop())

	case <-ctx.Done():
		return s.stop()
	}
}

func (s *Server) stop() error {
	s.log.Info("shutdown server...")
	shutdownCtx, cancel := context.WithTimeout(
		context.Background(), s.cfg.ShutdownTimeout(),
	)
	defer cancel()

	grpcErr := s.grpcStop(shutdownCtx)
	metricsErr := s.metricsStop(shutdownCtx)
	tracingErr := s.tracing.Shutdown(shutdownCtx)

	s.log.Warn("server stop")
	return errors.Join(grpcErr, metricsErr, tracingErr)
}

func (s *Server) grpcStop(ctx context.Context) error {
	stopped := make(chan struct{})

	go func() {
		s.grpc.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		return nil
	case <-ctx.Done():
		s.grpc.Stop()
		return fmt.Errorf("shutdown grpc server: timeout %s exceeded", s.cfg.ShutdownTimeout())
	}
}

func (s *Server) metricsStop(ctx context.Context) error {
	if s.metricsServer == nil {
		return nil
	}

	if err := s.metricsServer.Shutdown(ctx); err != nil {
		_ = s.metricsServer.Close()
		return fmt.Errorf("shutdown metrics server: timeout %s exceeded", s.cfg.ShutdownTimeout())
	}

	return nil
}
