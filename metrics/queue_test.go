package metrics

import (
	"reflect"
	"testing"

	"github.com/pkg/errors"

	"github.com/negz/q"
	"github.com/negz/q/e"
	"github.com/negz/q/test/fixtures"
)

func TestMetrics(t *testing.T) {
	t.Run("Add", func(t *testing.T) {
		msg := q.NewMessage([]byte("add"))
		queue := Queue(fixtures.NewPredictableQueue(nil, nil), NewNop())
		if err := queue.Add(msg); err != nil {
			t.Errorf("queue.Add(%v): %v", msg, err)
		}
	})

	t.Run("AddFull", func(t *testing.T) {
		msg := q.NewMessage([]byte("add"))
		queue := Queue(fixtures.NewPredictableQueue(nil, e.ErrFull(errors.New("full!"))), NewNop())
		if err := queue.Add(msg); !e.IsFull(err) {
			t.Errorf("queue.Add(%v): want error satisfying e.IsFull(), got %v", msg, err)
		}
	})

	t.Run("Peek", func(t *testing.T) {
		msg := q.NewMessage([]byte("peek"))
		queue := Queue(fixtures.NewPredictableQueue(msg, nil), NewNop())
		m, err := queue.Peek()
		if err != nil {
			t.Errorf("queue.Peek(): %v", err)
		}
		if !reflect.DeepEqual(msg, m) {
			t.Errorf("queue.Peek(): want %v, got %v", msg, m)
		}
	})

	t.Run("PeekEmpty", func(t *testing.T) {
		queue := Queue(fixtures.NewPredictableQueue(nil, e.ErrNotFound(errors.New("empty!"))), NewNop())
		if _, err := queue.Peek(); !e.IsNotFound(err) {
			t.Errorf("queue.Peek(): want error satisfying e.IsNotFound(), got %v", err)
		}
	})

	t.Run("Pop", func(t *testing.T) {
		msg := q.NewMessage([]byte("pop"))
		queue := Queue(fixtures.NewPredictableQueue(msg, nil), NewNop())
		m, err := queue.Pop()
		if err != nil {
			t.Errorf("queue.Pop(): %v", err)
		}
		if !reflect.DeepEqual(msg, m) {
			t.Errorf("queue.Pop(): want %v, got %v", msg, m)
		}
	})

	t.Run("PopEmpty", func(t *testing.T) {
		queue := Queue(fixtures.NewPredictableQueue(nil, e.ErrNotFound(errors.New("empty!"))), NewNop())
		if _, err := queue.Pop(); !e.IsNotFound(err) {
			t.Errorf("queue.Pop(): want error satisfying e.IsNotFound(), got %v", err)
		}
	})
}
