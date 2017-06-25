package fixtures

import (
	"time"

	"github.com/google/uuid"

	"github.com/negz/q"
)

type predictableQueue struct {
	err error
	msg *q.Message
}

// NewPredictableQueue returns a queue that  always returns the error and/or
// message provided.
func NewPredictableQueue(m *q.Message, err error) q.Queue {
	return &predictableQueue{err: err, msg: m}
}

func (p *predictableQueue) ID() uuid.UUID {
	return uuid.Must(uuid.Parse("92082756-edea-48ca-9cf0-870a9b1fa2eb"))
}

func (p *predictableQueue) Store() q.Store {
	return q.Memory
}

func (p *predictableQueue) Created() time.Time {
	return time.Unix(0, 0)
}

func (p *predictableQueue) Tags() *q.Tags {
	t := &q.Tags{}
	t.Add("log", "captain")
	t.Add("log", "stardate 42073.1")
	return t
}

func (p *predictableQueue) Add(m *q.Message) error {
	return p.err
}

func (p *predictableQueue) Pop() (*q.Message, error) {
	return p.msg, p.err
}

func (p *predictableQueue) Peek() (*q.Message, error) {
	return p.msg, p.err
}
