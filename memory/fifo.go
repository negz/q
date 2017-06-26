// Package memory provides an in-memory FIFO queue backed by a linked list.
package memory

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/negz/q"
	"github.com/negz/q/e"
)

type fifo struct {
	meta  *q.Metadata
	ll    *linkedList
	limit int
	m     *sync.RWMutex
}

// An Option represents an optional argument to a new in-memory FIFO queue.
type Option func(*fifo)

// Limit specifies the maximum number of messages that may exist in a queue.
// Unbounded queues will accept messages until they exhaust available resources.
func Limit(l int) Option {
	return func(f *fifo) {
		f.limit = l
	}
}

// Tagged applies the provided tags to a new queue.
func Tagged(t ...q.Tag) Option {
	return func(f *fifo) {
		for _, tag := range t {
			f.meta.Tags.AddTag(tag)
		}
	}
}

// New returns a new FIFO queue backed by an in-memory linked list.
func New(o ...Option) q.Queue {
	meta := &q.Metadata{ID: uuid.New(), Created: time.Now(), Tags: &q.Tags{}}
	f := &fifo{meta: meta, ll: &linkedList{}, limit: q.Unbounded, m: &sync.RWMutex{}}
	for _, opt := range o {
		opt(f)
	}
	return f
}

func (f *fifo) ID() uuid.UUID {
	return f.meta.ID
}

func (f *fifo) Store() q.Store {
	return q.Memory
}

func (f *fifo) Created() time.Time {
	return f.meta.Created
}

func (f *fifo) Tags() *q.Tags {
	return f.meta.Tags
}

func (f *fifo) Add(m *q.Message) error {
	f.m.Lock()
	defer f.m.Unlock()
	if (f.limit != q.Unbounded) && (f.ll.length >= f.limit) {
		return e.ErrFull(errors.Errorf("queue %s has reached limit of %d messages", f.ID(), f.limit))
	}
	f.ll.add(m)
	return nil
}

func (f *fifo) Pop() (*q.Message, error) {
	f.m.Lock()
	defer f.m.Unlock()
	m := f.ll.pop()
	if m == nil {
		return nil, e.ErrNotFound(errors.Errorf("queue %s is empty", f.ID()))
	}
	return m, nil
}

func (f *fifo) Peek() (*q.Message, error) {
	f.m.RLock()
	defer f.m.RUnlock()
	m := f.ll.peek()
	if m == nil {
		return nil, e.ErrNotFound(errors.Errorf("queue %s is empty", f.ID()))
	}
	return m, nil
}
