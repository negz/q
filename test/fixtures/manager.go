package fixtures

import (
	"github.com/google/uuid"
	"github.com/negz/q"
)

type predictableManager struct {
	q   q.Queue
	err error
}

// NewPredictableManager returns a manager that  always returns the error and/or
// queue provided.
func NewPredictableManager(queue q.Queue, err error) q.Manager {
	return &predictableManager{q: queue, err: err}
}

func (m *predictableManager) Add(queue q.Queue) error {
	return m.err
}

func (m *predictableManager) Get(id uuid.UUID) (q.Queue, error) {
	return m.q, m.err
}

func (m *predictableManager) Delete(id uuid.UUID) error {
	return m.err
}

func (m *predictableManager) List() ([]q.Queue, error) {
	return []q.Queue{m.q}, m.err
}
