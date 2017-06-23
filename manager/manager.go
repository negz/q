package manager

import (
	"sync"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/negz/q"
	"github.com/negz/q/e"
)

type manager struct {
	m  map[uuid.UUID]q.Queue
	mx *sync.RWMutex
}

// New returns a new in-memory queue manager.
func New() q.Manager {
	return &manager{m: make(map[uuid.UUID]q.Queue), mx: &sync.RWMutex{}}
}

func (m *manager) Add(queue q.Queue) error {
	m.mx.Lock()
	defer m.mx.Unlock()
	m.m[queue.ID()] = queue
	// An add to a map will always succeed, but our interface supports
	// returning an error for future compatibility with more complex backing
	// stores.
	return nil
}

func (m *manager) Get(id uuid.UUID) (q.Queue, error) {
	m.mx.RLock()
	defer m.mx.RUnlock()
	queue, ok := m.m[id]
	if !ok {
		return nil, e.ErrNotFound(errors.Errorf("cannot find queue with id %s", id))
	}
	return queue, nil
}

func (m *manager) Delete(id uuid.UUID) error {
	m.mx.Lock()
	defer m.mx.Unlock()
	delete(m.m, id)
	// A delete from a map will always succeed, but our interface supports
	// returning an error for future compatibility with more complex backing
	// stores.
	return nil
}

func (m *manager) List() ([]q.Queue, error) {
	m.mx.RLock()
	defer m.mx.RUnlock()
	l := make([]q.Queue, 0, len(m.m))
	for _, queue := range m.m {
		l = append(l, queue)
	}
	// A read from a map will always succeed, but our interface supports
	// returning an error for future compatibility with more complex backing
	// stores.
	return l, nil
}
