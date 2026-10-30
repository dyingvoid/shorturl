package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	grpc          *grpc.Server
	metricsServer *http.Server

	cfg Config
	
	log *slog.Logger
	id  string
}

type Config interface {
	Port() int
	MetricsPort() int
	ShutdownTimeout() time.Duration
}

func New(
	cfg Config,
	log *slog.Logger,
	id string,
	reg prometheus.Gatherer,
) *Server {
	log = log.With("server_id", id)

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			RequestID(),
			Logger(log),
			Trace(),
			Recovery(log),
		),
	)
	reflection.Register(grpcServer)

	server := &Server{
		grpc: grpcServer,
		cfg:  cfg,
		log:  log,
		id:   id,
	}

	if reg != nil {
		server.metricsServer = &http.Server{
			Addr: fmt.Sprintf(":%d", cfg.MetricsPort()),
			Handler: promhttp.HandlerFor(
				reg,
				promhttp.HandlerOpts{},
			),
		}
	}

	return server
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
		defer close(metricsErrCh)
		if err := s.metricsServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			metricsErrCh <- err
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
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			s.cfg.ShutdownTimeout(),
		)
		defer cancel()

		metricsError := s.metricsStop(shutdownCtx)
		return errors.Join(grpcErr, metricsError)

	case metricsErr := <-metricsErrCh:
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			s.cfg.ShutdownTimeout(),
		)
		defer cancel()

		grpcErr := s.grpcStop(shutdownCtx)
		return errors.Join(metricsErr, grpcErr)

	case <-ctx.Done():
		s.log.Warn("shutdown server...")
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(), s.cfg.ShutdownTimeout(),
		)
		defer cancel()

		grpcErr := s.grpcStop(shutdownCtx)
		metricsErr := s.metricsStop(shutdownCtx)

		s.log.Warn("server stop")
		return errors.Join(grpcErr, metricsErr)
	}
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
	if err := s.metricsServer.Shutdown(ctx); err != nil {
		_ = s.metricsServer.Close()
		return fmt.Errorf("shutdown metrics server: timeout %s exceeded", s.cfg.ShutdownTimeout())
	}

	return nil
}
