package metrics

import (
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/negz/q"
)

type prom struct {
	enqueued *prometheus.CounterVec
	consumed *prometheus.CounterVec
	errors   *prometheus.CounterVec
}

// NewPrometheus returns a new implementation of Metrics that exposes metrics to
// Prometheus.
func NewPrometheus(r prometheus.Registerer) (q.Metrics, prometheus.Gatherer) {
	enqueued := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "queue_messages_enqueued_total",
			Help: "Number of queued messages.",
		},
		[]string{"queue"},
	)
	consumed := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "queue_messages_consumed_total",
			Help: "Number of consumed messages.",
		},
		[]string{"queue"},
	)
	errors := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "queue_errors_total",
			Help: "Number of errors encountered while enqueuing or consuming messages.",
		},
		[]string{"queue", "type"},
	)

	r.MustRegister(prometheus.NewGoCollector())
	r.MustRegister(prometheus.NewProcessCollector(os.Getpid(), ""))
	r.MustRegister(enqueued)
	r.MustRegister(consumed)
	r.MustRegister(errors)

	return &prom{enqueued, consumed, errors}, r
}

func (m *prom) Enqueued(id uuid.UUID) {
	m.enqueued.With(prometheus.Labels{"queue": fmt.Sprint(id)}).Inc()
}

func (m *prom) Consumed(id uuid.UUID) {
	m.consumed.With(prometheus.Labels{"queue": fmt.Sprint(id)}).Inc()
}

func (m *prom) Error(id uuid.UUID, t q.Error) {
	labels := prometheus.Labels{
		"queue": fmt.Sprint(id),
		"type":  fmt.Sprint(t),
	}
	m.errors.With(labels).Inc()
}
