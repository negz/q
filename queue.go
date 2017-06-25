package q

import (
	"time"

	"github.com/google/uuid"
)

// Unbounded queues will accept messages until they exhaust available resources.
const Unbounded int = -1

// A Store is the type of backing store a queue uses.
type Store int

const (
	// UnknownStore queues have an indeterminate backing store.
	UnknownStore Store = iota

	// Memory queues are in-memory. Their contents do not persist across process restarts.
	Memory
)

// Error differentiates errors for metric collection purposes.
type Error int

const (
	// UnknownError indicates an unknown error type.
	UnknownError Error = iota

	// Full indicates a queue was full.
	Full

	// NotFound indicates no messages were found in a queue.
	NotFound
)

// Metadata is useful information associated with either queues or messages.
type Metadata struct {
	ID      uuid.UUID // ID is a globally unique identifier for a resource.
	Created time.Time // Created is the creation time of a resource.
	Tags    *Tags     // Tags are arbitrary key:value pairs associated with a resource.
}

// A Message represents an entry in a queue.
type Message struct {
	*Metadata
	Payload []byte // The Payload of a Message is an arbitrary byte array.
}

// An Option represents an optional argument to a new message.
type Option func(*Message)

// Tagged applies the provided tags to a new message.
func Tagged(t ...Tag) Option {
	return func(m *Message) {
		for _, tag := range t {
			m.Tags.AddTag(tag)
		}
	}
}

// NewMessage creates a message from the supplied payload.
func NewMessage(payload []byte, o ...Option) *Message {
	m := &Message{Metadata: &Metadata{ID: uuid.New(), Created: time.Now(), Tags: &Tags{}}, Payload: payload}
	for _, opt := range o {
		opt(m)
	}
	return m
}

// A Queue stores Messages for consumption by another process.
type Queue interface {
	ID() uuid.UUID           // ID is the globally unique identifier for this queue.
	Created() time.Time      // Created is the creation time of this queue.
	Tags() *Tags             // Tags are arbitrary key:value pairs associated with this queue.
	Store() Store            // Store indicates which backing store this queue uses.
	Add(*Message) error      // Add amends a message to this queue.
	Pop() (*Message, error)  // Pop consumes and returns the next message in the queue.
	Peek() (*Message, error) // Peek returns the next message in the queue without consuming it.
}

// Metrics for a queue.
// We only expose counts, not gauges, because they don't lose meaning when
// downsampled in a timeseries. See https://goo.gl/WTHgAq for details.
type Metrics interface {
	Enqueued(id uuid.UUID) // Enqueued increments the enqueued message count.
	Consumed(id uuid.UUID) // Consumed increments the consumed message count.
	// Error increments the count of errors encountered while queueing or consuming messages.
	Error(id uuid.UUID, t Error)
}

// A Manager manages a set of queues.
type Manager interface {
	Add(Queue) error                 // Add a new queue to the manager.
	Get(id uuid.UUID) (Queue, error) // Get an existing queue given its ID.
	Delete(id uuid.UUID) error       // Delete an existing queue given its ID.
	List() ([]Queue, error)          // List all existing queues.
}

// A Factory produces new queues with the requested store, limit, and tags.
type Factory interface {
	New(s Store, limit int, t ...Tag) (Queue, error)
}
