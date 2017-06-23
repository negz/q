package logging

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/negz/q"
	"github.com/negz/q/e"
	"github.com/negz/q/test/fixtures"
)

type predictableManager struct {
	q   q.Queue
	err error
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

func TestManager(t *testing.T) {
	t.Run("Add", func(t *testing.T) {
		m := Manager(&predictableManager{}, zap.NewNop())
		m.Add(fixtures.NewPredictableQueue(nil, nil))
	})

	t.Run("AddErr", func(t *testing.T) {
		queue := fixtures.NewPredictableQueue(nil, nil)
		want := errors.New("boom!")
		m := Manager(&predictableManager{err: want}, zap.NewNop())
		if err := m.Add(queue); err != want {
			t.Errorf("m.Add(%v): %v", queue, err)
		}
	})

	t.Run("Get", func(t *testing.T) {
		queue := fixtures.NewPredictableQueue(nil, nil)
		m := Manager(&predictableManager{q: queue}, zap.NewNop())
		got, err := m.Get(queue.ID())
		if err != nil {
			t.Errorf("m.Get(%v): %v", queue.ID(), err)
			return
		}

		if !reflect.DeepEqual(queue, got) {
			t.Errorf("m.Get(%v):\nwant %+#v\ngot %+#v", queue, got)
		}
	})

	t.Run("GetNotFound", func(t *testing.T) {
		queue := fixtures.NewPredictableQueue(nil, nil)
		m := Manager(&predictableManager{err: e.ErrNotFound(errors.New("not found!"))}, zap.NewNop())
		if _, err := m.Get(queue.ID()); !e.IsNotFound(err) {
			t.Errorf("m.Get(%v): %v", queue.ID(), err)
		}
	})

	t.Run("List", func(t *testing.T) {
		queue := fixtures.NewPredictableQueue(nil, nil)
		m := Manager(&predictableManager{q: queue}, zap.NewNop())
		want := []q.Queue{queue}
		got, err := m.List()
		if err != nil {
			t.Errorf("m.List(): %v", err)
			return
		}

		if !reflect.DeepEqual(want, got) {
			t.Errorf("m.List():\nwant %v\ngot %v", want, got)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		queue := fixtures.NewPredictableQueue(nil, nil)
		m := Manager(&predictableManager{q: queue}, zap.NewNop())
		if err := m.Delete(queue.ID()); err != nil {
			t.Errorf("m.Delete(%v): %v", queue.ID(), err)
		}
	})

	t.Run("DeleteErr", func(t *testing.T) {
		want := errors.New("boom!")
		m := Manager(&predictableManager{err: want}, zap.NewNop())
		id := uuid.New()
		if err := m.Delete(id); err != want {
			t.Errorf("m.Delete(%v): %v", id, err)
		}
	})
}
