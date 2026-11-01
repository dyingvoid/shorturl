package server

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type MetricsRegistry interface {
	prometheus.Registerer
	prometheus.Gatherer
}

type metrics struct {
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
}

func newMetrics(reg prometheus.Registerer) (*metrics, error) {
	m := &metrics{
		requests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "grpc_requests_total",
				Help: "Total number of gRPC requests.",
			},
			[]string{"method", "status"},
		),
		duration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "grpc_request_duration_seconds",
				Help:    "Duration of gRPC requests.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method"},
		),
	}

	if err := reg.Register(m.requests); err != nil {
		return nil, err
	}

	if err := reg.Register(m.duration); err != nil {
		reg.Unregister(m.requests)
		return nil, err
	}

	return m, nil
}

func (m *metrics) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		m.requests.
			WithLabelValues(info.FullMethod, status.Code(err).String()).
			Inc()

		m.duration.
			WithLabelValues(info.FullMethod).
			Observe(time.Since(start).Seconds())

		return resp, err
	}
}
