package infrastructure_prometheus

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	linksCreated  prometheus.Counter
	redirects     prometheus.Counter
	kafkaMessages prometheus.Counter
}

func NewMetrics(reg prometheus.Registerer) (*Metrics, error) {
	m := &Metrics{
		linksCreated: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "links_created_total",
			Help: "Total number of created short links.",
		}),
		redirects: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "redirects_total",
			Help: "Total number of successful redirects.",
		}),
		kafkaMessages: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "kafka_messages_produced_total",
			Help: "Total number of messages produced to Kafka.",
		}),
	}

	collectors := []prometheus.Collector{
		m.linksCreated,
		m.redirects,
		m.kafkaMessages,
	}

	for i, collector := range collectors {
		if err := reg.Register(collector); err != nil {
			for _, registered := range collectors[:i] {
				reg.Unregister(registered)
			}

			return nil, fmt.Errorf("register url metrics: %w", err)
		}
	}

	return m, nil
}

func (m *Metrics) IncLinksCreated() {
	m.linksCreated.Inc()
}

func (m *Metrics) IncRedirects() {
	m.redirects.Inc()
}

func (m *Metrics) IncKafkaMessagesProduced() {
	m.kafkaMessages.Inc()
}
