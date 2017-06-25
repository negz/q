package memory

import (
	"reflect"
	"testing"

	"github.com/negz/q"
)

var linkedListTests = []struct {
	list [][]byte
}{
	{list: [][]byte{[]byte{1}, []byte{2}, []byte{3}}},
	{list: [][]byte{[]byte{1}}},
	{list: [][]byte{}},
	{list: [][]byte{[]byte("satellites"), []byte("are"), []byte("cool")}},
}

func fromSlice(s [][]byte) *linkedList {
	l := &linkedList{}
	for _, b := range s {
		l.add(q.NewMessage(b))
	}
	return l
}

func toSlice(l *linkedList) [][]byte {
	s := make([][]byte, 0, l.length)
	e := l.head
	for e != nil {
		s = append(s, e.message.Payload)
		e = e.next
	}
	return s
}

func TestLinkedList(t *testing.T) {
	t.Run("add", func(t *testing.T) {
		for _, tt := range linkedListTests {
			ll := fromSlice(tt.list)
			got := toSlice(ll)
			if !reflect.DeepEqual(tt.list, got) {
				t.Errorf("want %v, got %v", tt.list, got)
			}
			if ll.length != len(tt.list) {
				t.Errorf("ll.length: want %v, got %v", len(tt.list), ll.length)
			}
		}

	})

	t.Run("peek", func(t *testing.T) {
		for _, tt := range linkedListTests {
			ll := fromSlice(tt.list)
			got := ll.peek()
			if len(tt.list) < 1 {
				if got != nil {
					t.Errorf("ll.peek(): want nil, got %v", got)
				}
				continue
			}
			if got == nil {
				t.Errorf("ll.peek(): want %v, got nil", tt.list[0])
				continue
			}
			if !reflect.DeepEqual(got.Payload, tt.list[0]) {
				t.Errorf("ll.peek(): want %v, got %v", tt.list[0], got.Payload)
			}
		}
	})

	t.Run("pop", func(t *testing.T) {
		for _, tt := range linkedListTests {
			ll := fromSlice(tt.list)
			got := ll.pop()
			if len(tt.list) < 1 {
				if got != nil {
					t.Errorf("ll.pop(): want nil, got %v", got)
				}
				continue
			}
			if got == nil {
				t.Errorf("ll.pop(): want %v, got nil", tt.list[0])
				continue
			}
			if !reflect.DeepEqual(got.Payload, tt.list[0]) {
				t.Errorf("ll.pop(): want %v, got %v", tt.list[0], got.Payload)
			}
		}
	})
}

// TestAddPopAddPop tests for a bug where new messages added to a linked list
// cannot be consumed if the queue has previously been populated and consumed.
func TestAddPopAddPop(t *testing.T) {
	for _, tt := range linkedListTests {
		l := &linkedList{}
		for _, b := range tt.list {
			l.add(&q.Message{Payload: b})
			if got := l.pop(); !reflect.DeepEqual(got.Payload, b) {
				t.Errorf("l.pop(): want %v, got %v", b, got.Payload)
				continue
			}
		}
	}
}
