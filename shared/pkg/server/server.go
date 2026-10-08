package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	grpc *grpc.Server
	cfg  Config
	log  *slog.Logger
}

type Config interface {
	Port() int
	ShutdownTimeout() time.Duration
}

func New(cfg Config, log *slog.Logger) *Server {
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			RequestID(),
			Logger(log),
			Trace(),
			Recovery(log),
		),
	)

	reflection.Register(grpcServer)

	return &Server{
		grpc: grpcServer,
		cfg:  cfg,
		log:  log,
	}
}

func (s *Server) GRPC() *grpc.Server {
	return s.grpc
}

func (s *Server) Run(ctx context.Context) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.cfg.Port()))
	if err != nil {
		return fmt.Errorf("listen grpc: %w", err)
	}

	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)

		s.log.Warn("server starting", "port", s.cfg.Port())

		err := s.grpc.Serve(lis)
		if err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("serve grpc: %w", err)
		}
		return nil
	case <-ctx.Done():
		s.log.Warn("shutdown grpc server...")

		if err := s.stop(); err != nil {
			return err
		}

		s.log.Warn("grpc server stopped")
		return nil
	}
}

func (s *Server) stop() error {
	stopped := make(chan struct{})

	go func() {
		s.grpc.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		return nil
	case <-time.After(s.cfg.ShutdownTimeout()):
		s.grpc.Stop()
		return fmt.Errorf("shutdown grpc server: timeout %s exceeded", s.cfg.ShutdownTimeout())
	}
}
