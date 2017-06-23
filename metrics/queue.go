package metrics

import (
	"time"

	"github.com/google/uuid"

	"github.com/negz/q"
	"github.com/negz/q/e"
)

type queue struct {
	w q.Queue
	m q.Metrics
}

// Queue wraps a queue with the supplied metrics.
func Queue(wrap q.Queue, m q.Metrics) q.Queue {
	return &queue{w: wrap, m: m}
}

func (l *queue) ID() uuid.UUID {
	return l.w.ID()
}

func (l *queue) Store() q.Store {
	return l.w.Store()
}

func (l *queue) Created() time.Time {
	return l.w.Created()
}

func (l *queue) Tags() *q.Tags {
	return l.w.Tags()
}

func (l *queue) Add(m *q.Message) error {
	if err := l.w.Add(m); err != nil {
		t := q.UnknownError
		if e.IsFull(err) {
			t = q.Full
		}
		l.m.Error(l.ID(), t)
		return err
	}
	l.m.Enqueued(l.ID())
	return nil
}

func (l *queue) Pop() (*q.Message, error) {
	m, err := l.w.Pop()
	if err != nil {
		t := q.UnknownError
		if e.IsNotFound(err) {
			t = q.NotFound
		}
		l.m.Error(l.ID(), t)
		return nil, err
	}
	l.m.Consumed(l.ID())
	return m, nil
}

func (l *queue) Peek() (*q.Message, error) {
	m, err := l.w.Peek()
	if err != nil {
		t := q.UnknownError
		if e.IsNotFound(err) {
			t = q.NotFound
		}
		l.m.Error(l.ID(), t)
		return nil, err
	}
	return m, nil
}
