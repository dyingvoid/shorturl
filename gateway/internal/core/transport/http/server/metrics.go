package core_http_server

import (
	"net/http"
	"strconv"
	"time"

	core_http_middleware "github.com/dyingvoid/shorturl/gateway/internal/core/transport/http/middleware"
	core_http_response "github.com/dyingvoid/shorturl/gateway/internal/core/transport/http/response"
	"github.com/prometheus/client_golang/prometheus"
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
				Name: "http_requests_total",
				Help: "Total number of HTTP requests.",
			},
			[]string{"method", "path", "status"},
		),
		duration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Duration of HTTP requests.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path"},
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

func (m *metrics) Middleware() core_http_middleware.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := core_http_response.NewWriter(w)

			defer func() {
				m.requests.
					WithLabelValues(r.Method, r.URL.Path, strconv.Itoa(rw.Status())).
					Inc()

				m.duration.
					WithLabelValues(r.Method, r.URL.Path).
					Observe(time.Since(start).Seconds())
			}()

			next.ServeHTTP(rw, r)
		})
	}
}
