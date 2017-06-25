// Package factory provides a FIFO queue factory.
package factory

import (
	"github.com/pkg/errors"

	"github.com/negz/q"
	"github.com/negz/q/e"
	"github.com/negz/q/memory"
)

// Default is the default queue factory. It can currently only produce in-memory
// FIFO queues.
var Default = &defaultFactory{}

type defaultFactory struct{}

func (f *defaultFactory) New(s q.Store, limit int, t ...q.Tag) (q.Queue, error) {
	switch s {
	case q.Memory:
		return memory.New(memory.Limit(limit), memory.Tagged(t...)), nil
	default:
		return nil, e.ErrNotFound(errors.New("unknown store type"))
	}
}
