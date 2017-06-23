package logging

import (
	"reflect"
	"testing"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/negz/q"
	"github.com/negz/q/e"
	"github.com/negz/q/test/fixtures"
)

func TestLogging(t *testing.T) {
	t.Run("Add", func(t *testing.T) {
		msg := q.NewMessage([]byte("add"))
		q := Queue(fixtures.NewPredictableQueue(nil, nil), zap.NewNop())
		if err := q.Add(msg); err != nil {
			t.Errorf("q.Add(%v): %v", msg, err)
		}
	})

	t.Run("AddFull", func(t *testing.T) {
		msg := q.NewMessage([]byte("add"))
		q := Queue(fixtures.NewPredictableQueue(nil, e.ErrFull(errors.New("full!"))), zap.NewNop())
		if err := q.Add(msg); !e.IsFull(err) {
			t.Errorf("queue.Add(%v): want error satisfying e.IsFull(), got %v", msg, err)
		}
	})

	t.Run("Peek", func(t *testing.T) {
		msg := q.NewMessage([]byte("peek"))
		queue := Queue(fixtures.NewPredictableQueue(msg, nil), zap.NewNop())
		m, err := queue.Peek()
		if err != nil {
			t.Errorf("queue.Peek(): %v", err)
		}
		if !reflect.DeepEqual(msg, m) {
			t.Errorf("queue.Peek(): want %v, got %v", msg, m)
		}
	})

	t.Run("PeekNotFound", func(t *testing.T) {
		queue := Queue(fixtures.NewPredictableQueue(nil, e.ErrNotFound(errors.New("empty!"))), zap.NewNop())
		if _, err := queue.Peek(); !e.IsNotFound(err) {
			t.Errorf("queue.Peek(): want error satisfying e.IsNotFound(), got %v", err)
		}
	})

	t.Run("Pop", func(t *testing.T) {
		msg := q.NewMessage([]byte("pop"))
		queue := Queue(fixtures.NewPredictableQueue(msg, nil), zap.NewNop())
		m, err := queue.Pop()
		if err != nil {
			t.Errorf("queue.Pop(): %v", err)
		}
		if !reflect.DeepEqual(msg, m) {
			t.Errorf("queue.Pop(): want %v, got %v", msg, m)
		}
	})

	t.Run("PopNotFound", func(t *testing.T) {
		queue := Queue(fixtures.NewPredictableQueue(nil, e.ErrNotFound(errors.New("empty!"))), zap.NewNop())
		if _, err := queue.Pop(); !e.IsNotFound(err) {
			t.Errorf("queue.Pop(): want error satisfying e.IsNotFound(), got %v", err)
		}
	})
}
