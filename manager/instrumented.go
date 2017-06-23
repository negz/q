package manager

import (
	"go.uber.org/zap"

	"github.com/google/uuid"
	"github.com/negz/q"
	"github.com/negz/q/logging"
	"github.com/negz/q/metrics"
)

type instrumented struct {
	m   q.Manager
	mx  q.Metrics
	log *zap.Logger
}

// An Option represents an optional argument to an instrumented manager.
type Option func(*instrumented)

// WithLogger instruments a queue manager with a zap Logger.
func WithLogger(l *zap.Logger) Option {
	return func(m *instrumented) {
		m.log = l
	}
}

// WithMetrics instruments a queue manager with Metrics.
func WithMetrics(mx q.Metrics) Option {
	return func(m *instrumented) {
		m.mx = mx
	}
}

// Instrumented returns a queue manager optionally instrumented with logging
// and/or metrics for both the manager itself and the queues it manages.
func Instrumented(m q.Manager, o ...Option) q.Manager {
	i := &instrumented{m: m, mx: metrics.NewNop(), log: zap.NewNop()}
	for _, opt := range o {
		opt(i)
	}
	i.m = logging.Manager(i.m, i.log)
	return i
}

func (i *instrumented) Add(queue q.Queue) error {
	queue = metrics.Queue(queue, i.mx)
	queue = logging.Queue(queue, i.log)
	return i.m.Add(queue)
}

func (i *instrumented) Get(id uuid.UUID) (q.Queue, error) {
	return i.m.Get(id)
}

func (i *instrumented) Delete(id uuid.UUID) error {
	return i.m.Delete(id)
}

func (i *instrumented) List() ([]q.Queue, error) {
	return i.m.List()
}
