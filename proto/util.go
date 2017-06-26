package proto

import (
	"fmt"
	"time"

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
	q.BoltDB:       BOLTDB,
}

// ToStore maps protobuf generated store types to q.Store.
var ToStore = map[Queue_Store]q.Store{
	UNKNOWN: q.UnknownStore,
	MEMORY:  q.Memory,
	BOLTDB:  q.BoltDB,
}

// ParseID parses a string ID into a uuid.UUID.
func ParseID(id string) (uuid.UUID, error) {
	u, err := uuid.Parse(id)
	if err != nil {
		return uuid.UUID{}, e.ErrInvalid(errors.Wrapf(err, "cannot parse ID %s as a UUID", id))
	}
	return u, nil
}

// FromMeta converts q.Metadata to its protobuf generated equivalent.
func FromMeta(m *q.Metadata) (*Metadata, error) {
	t, err := ptypes.TimestampProto(m.Created)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse timestamp")
	}
	return &Metadata{Id: fmt.Sprint(m.ID), Created: t, Tags: FromTags(m.Tags.Get())}, nil
}

// ToMeta converts protobuf generated code into q.Metadata.
func ToMeta(m *Metadata) (*q.Metadata, error) {
	id, err := ParseID(m.GetId())
	if err != nil {
		return nil, e.ErrInvalid(errors.Wrap(err, "cannot parse metadata ID"))
	}
	created := time.Unix(m.GetCreated().GetSeconds(), int64(m.GetCreated().GetNanos()))
	meta := &q.Metadata{ID: id, Created: created, Tags: &q.Tags{}}

	// TODO(negz): Revisit the Tags API. It's starting to feel pretty awkward.
	for _, t := range ToTags(m.GetTags()) {
		meta.Tags.AddTag(t)
	}
	return meta, nil
}

// FromQueue converts a q.Queue to its protobuf generated equivalent.
func FromQueue(queue q.Queue) (*Queue, error) {
	t, err := ptypes.TimestampProto(queue.Created())
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse timestamp")
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
		return nil, errors.Wrap(err, "cannot parse timestamp")
	}
	return &Message{
		Meta:    &Metadata{Id: fmt.Sprint(m.ID), Created: t, Tags: FromTags(m.Tags.Get())},
		Payload: m.Payload,
	}, nil
}

// ToMessage converts protobuf generated code into a *q.Message
func ToMessage(m *Message) (*q.Message, error) {
	meta, err := ToMeta(m.GetMeta())
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse metadata")
	}
	return &q.Message{Metadata: meta, Payload: m.GetPayload()}, nil
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
