package rpc

import (
	"fmt"

	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/negz/q"
	"github.com/negz/q/e"
	"github.com/negz/q/rpc/proto"
)

var storeToProto = map[q.Store]proto.Queue_Store{
	q.UnknownStore: proto.UNKNOWN,
	q.Memory:       proto.MEMORY,
}

var storeFromProto = map[proto.Queue_Store]q.Store{
	proto.UNKNOWN: q.UnknownStore,
	proto.MEMORY:  q.Memory,
}

func parseID(id string) (uuid.UUID, error) {
	u, err := uuid.Parse(id)
	if err != nil {
		return uuid.UUID{}, e.ErrInvalid(errors.Wrapf(err, "cannot parse ID %s as a UUID", id))
	}
	return u, nil
}

func queueToProto(queue q.Queue) (*proto.Queue, error) {
	t, err := ptypes.TimestampProto(queue.Created())
	if err != nil {
		return nil, e.ErrInvalid(errors.Wrap(err, "cannot parse timestamp"))
	}
	return &proto.Queue{
		Meta: &proto.Metadata{
			Id:      fmt.Sprint(queue.ID()),
			Created: t,
			Tags:    tagsToProto(queue.Tags().Get()),
		},
		Store: storeToProto[queue.Store()],
	}, nil
}

func messageToProto(m *q.Message) (*proto.Message, error) {
	t, err := ptypes.TimestampProto(m.Created)
	if err != nil {
		return nil, e.ErrInvalid(errors.Wrap(err, "cannot parse timestamp"))
	}
	return &proto.Message{
		Meta: &proto.Metadata{
			Id:      fmt.Sprint(m.ID),
			Created: t,
			Tags:    tagsToProto(m.Tags.Get()),
		},
		Payload: m.Payload,
	}, nil
}

func tagsToProto(t []q.Tag) []*proto.Tag {
	tags := make([]*proto.Tag, 0, len(t))
	for _, tag := range t {
		p := proto.Tag(tag)
		tags = append(tags, &p)
	}
	return tags
}

func tagsFromProto(t []*proto.Tag) []q.Tag {
	tags := make([]q.Tag, 0, len(t))
	for _, tag := range t {
		qt := q.Tag(*tag)
		tags = append(tags, qt)
	}
	return tags
}
