package proto

import (
	"fmt"

	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/negz/q"
	"github.com/negz/q/e"
)

// FromStore maps q.Store to its protobuf generated equivalent.
var FromStore = map[q.Store]Queue_Store{
	q.UnknownStore: UNKNOWN,
	q.Memory:       MEMORY,
}

// ToStore maps protobuf generated store types to q.Store.
var ToStore = map[Queue_Store]q.Store{
	UNKNOWN: q.UnknownStore,
	MEMORY:  q.Memory,
}

// ParseID parses a string ID into a uuid.UUID.
func ParseID(id string) (uuid.UUID, error) {
	u, err := uuid.Parse(id)
	if err != nil {
		return uuid.UUID{}, e.ErrInvalid(errors.Wrapf(err, "cannot parse ID %s as a UUID", id))
	}
	return u, nil
}

// FromQueue converts a q.Queue to its protobuf generated equivalent.
func FromQueue(queue q.Queue) (*Queue, error) {
	t, err := ptypes.TimestampProto(queue.Created())
	if err != nil {
		return nil, e.ErrInvalid(errors.Wrap(err, "cannot parse timestamp"))
	}
	return &Queue{
		Meta:  &Metadata{Id: fmt.Sprint(queue.ID()), Created: t, Tags: FromTags(queue.Tags().Get())},
		Store: FromStore[queue.Store()],
	}, nil
}

// FromMessage converts a *q.Message to its protobuf generated equivalent.
func FromMessage(m *q.Message) (*Message, error) {
	t, err := ptypes.TimestampProto(m.Created)
	if err != nil {
		return nil, e.ErrInvalid(errors.Wrap(err, "cannot parse timestamp"))
	}
	return &Message{
		Meta:    &Metadata{Id: fmt.Sprint(m.ID), Created: t, Tags: FromTags(m.Tags.Get())},
		Payload: m.Payload,
	}, nil
}

// FromTags converts q.Tag to its protobuf generated equivalent.
func FromTags(t []q.Tag) []*Tag {
	tags := make([]*Tag, 0, len(t))
	for _, tag := range t {
		p := Tag(tag)
		tags = append(tags, &p)
	}
	return tags
}

// ToTags converts protobuf generated code to a slice of q.Tag.
func ToTags(t []*Tag) []q.Tag {
	tags := make([]q.Tag, 0, len(t))
	for _, tag := range t {
		qt := q.Tag(*tag)
		tags = append(tags, qt)
	}
	return tags
}
