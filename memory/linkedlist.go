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
	if l.head == nil { // This list is empty.
		l.head = e
		l.tail = e
		l.length = 1
		return
	}
	l.tail.next = e
	l.tail = e
	l.length++
	return
}

func (l *linkedList) pop() *q.Message {
	if l.head == nil { // This list is empty.
		return nil
	}
	m := l.head.message
	l.length--
	if l.head.next == nil { // This list has a single element.
		l.head = nil
		l.tail = nil
		return m
	}
	l.head = l.head.next
	return m
}

func (l *linkedList) peek() *q.Message {
	if l.head == nil {
		return nil
	}
	return l.head.message
}
