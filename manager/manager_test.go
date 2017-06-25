package manager

import (
	"reflect"
	"testing"

	"github.com/negz/q"
	"github.com/negz/q/e"
	"github.com/negz/q/memory"
)

var managerTests = []struct {
	queues []q.Queue
}{
	{
		queues: []q.Queue{
			memory.New(memory.Limit(100), memory.Tagged(q.Tag{"position", "CAPCOM"})),
			memory.New(memory.Tagged(q.Tag{"position", "FLIGHT"})),
			memory.New(),
		},
	},
}

func TestManager(t *testing.T) {
	for _, tt := range managerTests {
		m := New()

		t.Run("Add", func(t *testing.T) {
			for _, queue := range tt.queues {
				m.Add(queue)
			}
		})

		t.Run("Get", func(t *testing.T) {
			for _, queue := range tt.queues {
				got, err := m.Get(queue.ID())
				if err != nil {
					t.Errorf("m.Get(%v): %v", queue.ID(), err)
					continue
				}

				if !reflect.DeepEqual(queue, got) {
					t.Errorf("m.Get(%v):\nwant %+#v\ngot %+#v", queue.ID(), queue, got)
				}
			}
		})

		t.Run("List", func(t *testing.T) {
			want := make(map[q.Queue]bool)
			got := make(map[q.Queue]bool)

			for _, queue := range tt.queues {
				want[queue] = true
			}

			l, _ := m.List()
			for _, queue := range l {
				got[queue] = true
			}

			// Compare the two slices of queues after converting them to 'sets'.
			// This allows to us to test that they have the same membership
			// without concern for their order.
			if !reflect.DeepEqual(want, got) {
				t.Errorf("m.List():\nwant %v\ngot %v", want, got)
			}
		})

		t.Run("Delete", func(t *testing.T) {
			for _, queue := range tt.queues {
				m.Delete(queue.ID())
				_, err := m.Get(queue.ID())
				if !e.IsNotFound(err) {
					t.Errorf("m.Get(%v): want error satisfying notFound(), got %v", queue.ID(), err)
				}
			}
		})
	}
}
