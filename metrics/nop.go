package metrics

import (
	"github.com/google/uuid"

	"github.com/negz/q"
)

type nopMetrics struct{}

// NewNop returns a metrics implementation that does nothing.
func NewNop() q.Metrics { return &nopMetrics{} }

func (m *nopMetrics) Enqueued(id uuid.UUID)         {}
func (m *nopMetrics) Consumed(id uuid.UUID)         {}
func (m *nopMetrics) Error(id uuid.UUID, t q.Error) {}
