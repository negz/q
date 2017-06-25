// Package logging provides logging wrappers for queues and queue managers.
package logging

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/google/uuid"
	"github.com/negz/q"
)

type manager struct {
	w   q.Manager
	log *zap.Logger
}

// Manager wraps a queue manager with the supplied logger.
func Manager(wrap q.Manager, l *zap.Logger) q.Manager {
	l.Debug("queue manager logging enabled")
	return &manager{w: wrap, log: l}
}

func (l *manager) Add(queue q.Queue) error {
	if err := l.w.Add(queue); err != nil {
		l.log.Error("add queue", idField(queue.ID()), zap.Error(err))
		return err
	}
	log := l.log
	for _, tag := range queue.Tags().Get() {
		log = log.With(zap.String("tag", fmt.Sprint(tag)))
	}
	log.Debug("add queue",
		idField(queue.ID()),
		zap.Time("created", queue.Created()))
	return nil
}

func (l *manager) Get(id uuid.UUID) (q.Queue, error) {
	queue, err := l.w.Get(id)
	if err != nil {
		l.log.Error("get queue", idField(id), zap.Error(err))
		return nil, err
	}
	l.log.Debug("get queue", idField(id))
	return queue, nil
}

func (l *manager) Delete(id uuid.UUID) error {
	if err := l.w.Delete(id); err != nil {
		l.log.Error("delete queue", idField(id), zap.Error(err))
		return err
	}
	l.log.Debug("delete queue", idField(id))
	return nil
}

func (l *manager) List() ([]q.Queue, error) {
	q, err := l.w.List()
	if err != nil {
		l.log.Error("list queues", zap.Error(err))
		return nil, err
	}
	l.log.Debug("list queues")
	return q, nil
}

func idField(id uuid.UUID) zapcore.Field {
	return zap.String("id", fmt.Sprint(id))
}
