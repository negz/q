package memory

import "github.com/negz/q"

type element struct {
	message *q.Message
	next    *element
}

/* linkedList exists mostly to demonstrate that I can implement a linked list.
   In practice I might use https://golang.org/src/container/list/list.go to
   avoid reinventing wheels. That said, a FIFO q does not require a doubly
   linked list, and this implementation has the advantage of storing a concrete
   type (*q.Message) rather than the empty interface.
*/
type linkedList struct {
	head   *element
	tail   *element
	length int
}

func (l *linkedList) add(m *q.Message) {
	e := &element{message: m}
	if l.tail == nil {
		l.head = e
		l.tail = e
		l.length = 1
		return
	}
	l.tail.next = e
	l.tail = l.tail.next
	l.length++
}

func (l *linkedList) pop() *q.Message {
	if l.head == nil {
		return nil
	}
	m := l.head.message
	l.head = l.head.next
	l.length--
	return m
}

func (l *linkedList) peek() *q.Message {
	if l.head == nil {
		return nil
	}
	return l.head.message
}
