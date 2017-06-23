package q

import (
	"fmt"
	"sync"
)

// A Tag is an arbitrary key:value pair associate with a resource.
type Tag struct {
	Key   string
	Value string
}

// String represents a tag as a string.
func (t Tag) String() string {
	return fmt.Sprintf("%s:%s", t.Key, t.Value)
}

// Tags are a threadsafe set of tags. A single key may have multiple values.
type Tags struct {
	t map[Tag]bool
	m *sync.RWMutex
}

func (t *Tags) init() {
	t.t = make(map[Tag]bool)
	t.m = &sync.RWMutex{}
}

// Get all tags.
func (t *Tags) Get() []Tag {
	if t.t == nil {
		return nil
	}
	t.m.RLock()
	defer t.m.RUnlock()
	l := make([]Tag, 0, len(t.t))
	for tag := range t.t {
		l = append(l, tag)
	}
	return l
}

// Contains indicates whether the given key value pair exists in a set of tags.
func (t *Tags) Contains(k, v string) bool {
	return t.ContainsTag(Tag{k, v})
}

// ContainsTag indicates whether the given tag exists in a set of tags.
func (t *Tags) ContainsTag(tag Tag) bool {
	if t.t == nil {
		return false
	}
	t.m.RLock()
	defer t.m.RUnlock()
	return t.t[tag]
}

// Add the supplied key value pair to a set of Tags.
func (t *Tags) Add(k, v string) {
	t.AddTag(Tag{k, v})
}

// AddTag adds the supplied Tag to a set of Tags.
func (t *Tags) AddTag(tag Tag) {
	if t.t == nil {
		t.init()
	}
	t.m.Lock()
	defer t.m.Unlock()
	t.t[tag] = true
}

// AddMap adds each key value pair in a map to a set of Tags.
func (t *Tags) AddMap(m map[string]string) {
	if t.t == nil {
		t.init()
	}
	t.m.Lock()
	defer t.m.Unlock()
	for k, v := range m {
		t.t[Tag{Key: k, Value: v}] = true
	}
}

// Remove the supplied key value pair from a set of Tags.
func (t *Tags) Remove(k, v string) {
	t.RemoveTag(Tag{k, v})
}

// RemoveTag removes the supplied tag from a set of Tags.
func (t *Tags) RemoveTag(tag Tag) {
	if t.t == nil {
		return
	}
	t.m.Lock()
	defer t.m.Unlock()
	delete(t.t, tag)
}
