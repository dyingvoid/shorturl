package infrastructure_prometheus

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	kafkaMessages       prometheus.Counter
	batchInsertDuration prometheus.Histogram
}

func NewMetrics(reg prometheus.Registerer) (*Metrics, error) {
	m := &Metrics{
		kafkaMessages: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "kafka_messages_consumed_total",
			Help: "Total number of messages consumed from Kafka.",
		}),
		batchInsertDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "clickhouse_batch_insert_duration_seconds",
			Help:    "Duration of ClickHouse batch inserts.",
			Buckets: prometheus.DefBuckets,
		}),
	}

	collectors := []prometheus.Collector{
		m.kafkaMessages,
		m.batchInsertDuration,
	}

	for i, collector := range collectors {
		if err := reg.Register(collector); err != nil {
			for _, registered := range collectors[:i] {
				reg.Unregister(registered)
			}

			return nil, fmt.Errorf("register analytics metrics: %w", err)
		}
	}

	return m, nil
}

func (m *Metrics) IncKafkaMessagesConsumed(count int) {
	m.kafkaMessages.Add(float64(count))
}

func (m *Metrics) ObserveClickHouseBatchInsertDuration(seconds float64) {
	m.batchInsertDuration.Observe(seconds)
}
