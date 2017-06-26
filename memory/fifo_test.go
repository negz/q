package memory

import (
	"reflect"
	"testing"

	"github.com/negz/q"
	"github.com/negz/q/e"
)

var fifoTests = []struct {
	messages []*q.Message
	limit    int
}{
	{
		messages: []*q.Message{
			q.NewMessage([]byte("salyut"), q.Tagged(q.Tag{"country", "USSR"})),
			q.NewMessage([]byte("DOS")),
			q.NewMessage([]byte("kosmos")),
			q.NewMessage([]byte("skylab")),
			q.NewMessage([]byte("mir")),
			q.NewMessage([]byte("iss")),
			q.NewMessage([]byte("tiangong")),
		},
		limit: q.Unbounded,
	},
	{
		messages: []*q.Message{
			q.NewMessage([]byte("salyut")),
			q.NewMessage([]byte("DOS")),
			q.NewMessage([]byte("kosmos")),
		},
		limit: 2,
	},
	{
		messages: []*q.Message{},
		limit:    q.Unbounded,
	},
}

func TestFIFO(t *testing.T) {
	for _, tt := range fifoTests {
		queue := New(Limit(tt.limit), Tagged(q.Tag{"function", "space station"}))

		t.Run("Add", func(t *testing.T) {
			for _, m := range tt.messages {
				if err := queue.Add(m); err != nil {
					if len(tt.messages) > tt.limit && e.IsFull(err) {
						continue
					}
					t.Errorf("queue.Add(%v): %v", m, err)
				}
			}
		})

		t.Run("Peek", func(t *testing.T) {
			m, err := queue.Peek()
			if err != nil {
				if len(tt.messages) < 1 && e.IsNotFound(err) {
					return
				}
				t.Errorf("queue.Peek(): %v", err)
				return
			}
			if !reflect.DeepEqual(tt.messages[0], m) {
				t.Errorf("queue.Peek(): want %v, got %v", tt.messages[0], m)
			}
		})

		t.Run("Pop", func(t *testing.T) {
			for i := range tt.messages {
				m, err := queue.Pop()
				if err != nil {
					if i == len(tt.messages)-1 && e.IsNotFound(err) {
						continue
					}
					t.Errorf("queue.Pop(): %v", err)
					continue
				}
				if !reflect.DeepEqual(tt.messages[i], m) {
					t.Errorf("queue.Pop(): want %v, got %v", tt.messages[i], m)
				}
			}
		})
	}
}
