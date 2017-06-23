package logging

import (
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/negz/q"
)

type queue struct {
	w   q.Queue
	log *zap.Logger
}

// Queue wraps a queue with the supplied logger.
func Queue(wrap q.Queue, l *zap.Logger) q.Queue {
	log := l.With(idField(wrap.ID()))
	log.Debug("queue logging enabled")
	return &queue{w: wrap, log: log}
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
	log := l.log.With(idField(m.ID))
	if err := l.w.Add(m); err != nil {
		log.Error("add", zap.Error(err))
		return err
	}
	log.Debug("add")
	return nil
}

func (l *queue) Pop() (*q.Message, error) {
	m, err := l.w.Pop()
	if err != nil {
		l.log.Error("pop", zap.Error(err))
		return nil, err
	}
	l.log.Debug("pop", idField(m.ID))
	return m, nil
}

func (l *queue) Peek() (*q.Message, error) {
	m, err := l.w.Peek()
	if err != nil {
		l.log.Error("peek", zap.Error(err))
		return nil, err
	}
	l.log.Debug("peek", idField(m.ID))
	return m, nil
}
