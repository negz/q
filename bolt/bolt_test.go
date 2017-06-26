package bdb

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/boltdb/bolt"
	"github.com/negz/q"
	"github.com/negz/q/e"
)

var boltTests = []struct {
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

func TestBolt(t *testing.T) {
	for _, tt := range boltTests {
		tmp, err := ioutil.TempDir(".", "qtestbolt")
		if err != nil {
			t.Fatalf("ioutil.TempDir(): %v", err)
		}
		defer os.RemoveAll(tmp)

		path := filepath.Join(tmp, "db")
		opts := &bolt.Options{Timeout: 1 * time.Second}
		db, err := bolt.Open(path, 0600, opts)
		if err != nil {
			t.Fatalf("bolt.Open(%v, %v, %v): %v", path, 0600, opts, err)
		}
		defer db.Close()

		queue, err := New(db, Limit(tt.limit))
		if err != nil {
			t.Fatalf("New(%v, Limit(%v)): %v", db, tt.limit, err)
		}

		queue, err = Open(db, queue.ID())
		if err != nil {
			t.Fatalf("Open(%v, %v): %v", db, queue.ID(), tt.limit, err)
		}

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
