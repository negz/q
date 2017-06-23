package rpc

import (
	"fmt"

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
	return u, e.ErrInvalid(errors.Wrap(err, "cannot parse UUID"))
}

func queueToProto(queue q.Queue) *proto.Queue {
	c := queue.Created()
	return &proto.Queue{
		Meta: &proto.Metadata{
			Id:      fmt.Sprint(queue.ID()),
			Created: &c,
			Tags:    tagsToProto(queue.Tags().Get()),
		},
		Store: storeToProto[queue.Store()],
	}
}

func messageToProto(m *q.Message) *proto.Message {
	return &proto.Message{
		Meta: &proto.Metadata{
			Id:      fmt.Sprint(m.ID),
			Created: &m.Created,
			Tags:    tagsToProto(m.Tags.Get()),
		},
	}
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
		tags = append(tags, q.Tag{tag.Key, tag.Value})
	}
	return tags
}
