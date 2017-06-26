package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/gogo/protobuf/jsonpb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/negz/q/proto"
)

var (
	ctx        = context.Background()
	marshaller = &jsonpb.Marshaler{Indent: "  ", EmitDefaults: true}
)

func queueStores() []string {
	s := make([]string, 0, len(proto.Queue_Store_value))
	for value := range proto.Queue_Store_value {
		s = append(s, value)
	}
	return s
}

func main() {
	var (
		app    = kingpin.New(filepath.Base(os.Args[0]), "Queries and manages a queue server.").DefaultEnvars()
		server = app.Flag("server", "Address at which to query queue server.").Short('s').Default(":10002").String()

		listQueues = app.Command("list", "List of all queues.")

		getQueue   = app.Command("get", "Get details of a queue.")
		getQueueID = getQueue.Arg("id", "ID of queue.").String()

		deleteQueue   = app.Command("delete", "Delete a queue.")
		deleteQueueID = deleteQueue.Arg("id", "ID of queue.").String()

		newQueue      = app.Command("new", "Create a queue.")
		newQueueStore = newQueue.Arg("store", "Backing store for queue.").HintAction(queueStores).String()
		newQueueLimit = newQueue.Arg("limit", "Message limit of queue. -1 for unlimited.").Int64()
		newQueueTags  = newQueue.Flag("tag", "Tag to apply to queue.").Short('t').StringMap()

		addQueueTag      = app.Command("tag", "Tag a queue.")
		addQueueTagID    = addQueueTag.Arg("id", "ID of queue.").String()
		addQueueTagKey   = addQueueTag.Arg("key", "Tag key.").String()
		addQueueTagValue = addQueueTag.Arg("value", "Tag value.").String()

		deleteQueueTag      = app.Command("untag", "Untag a queue.")
		deleteQueueTagID    = deleteQueueTag.Arg("id", "ID of queue.").String()
		deleteQueueTagKey   = deleteQueueTag.Arg("key", "Tag key.").String()
		deleteQueueTagValue = deleteQueueTag.Arg("value", "Tag value.").String()

		addMessage      = app.Command("add", "Add a message to a queue. Message payload is read from stdin.")
		addMessageQueue = addMessage.Arg("id", "ID of queue in which to add message.").String()
		addMessageTags  = addMessage.Flag("tag", "Tag to apply to message.").Short('t').StringMap()

		popMessage      = app.Command("pop", "Consume a message from the queue.")
		popMessageQueue = popMessage.Arg("queue", "ID of queue from which to pop message.").String()

		peekMessage      = app.Command("peek", "Preview a message from the queue.")
		peekMessageQueue = peekMessage.Arg("queue", "ID of queue in which to peek at message.").String()
	)
	kp := kingpin.MustParse(app.Parse(os.Args[1:]))

	conn, err := grpc.Dial(*server, grpc.WithInsecure())
	kingpin.FatalIfError(err, "cannot dial server %s", *server)
	defer conn.Close()
	h := &handlers{proto.NewQClient(conn)}

	switch kp {
	case listQueues.FullCommand():
		h.listQueues()
	case getQueue.FullCommand():
		h.getQueue(*getQueueID)
	case deleteQueue.FullCommand():
		h.deleteQueue(*deleteQueueID)
	case newQueue.FullCommand():
		h.newQueue(*newQueueStore, *newQueueLimit, *newQueueTags)
	case addQueueTag.FullCommand():
		h.addQueueTag(*addQueueTagID, *addQueueTagKey, *addQueueTagValue)
	case deleteQueueTag.FullCommand():
		h.deleteQueueTag(*deleteQueueTagID, *deleteQueueTagKey, *deleteQueueTagValue)
	case addMessage.FullCommand():
		h.addMessage(*addMessageQueue, *addMessageTags)
	case popMessage.FullCommand():
		h.popMessage(*popMessageQueue)
	case peekMessage.FullCommand():
		h.peekMessage(*peekMessageQueue)
	}
}

type handlers struct {
	c proto.QClient
}

func (h *handlers) listQueues() {
	rsp, err := h.c.ListQueues(ctx, &proto.ListQueuesRequest{})
	kingpin.FatalIfError(err, "cannot list queues")
	j, err := marshaller.MarshalToString(rsp)
	kingpin.FatalIfError(err, "cannot marshal queues to JSON:\n%#v", rsp)
	fmt.Printf("%s\n", j)
}

func (h *handlers) getQueue(id string) {
	rsp, err := h.c.GetQueue(ctx, &proto.GetQueueRequest{QueueId: id})
	kingpin.FatalIfError(err, "cannot get queue")
	j, err := marshaller.MarshalToString(rsp)
	kingpin.FatalIfError(err, "cannot marshal queue to JSON:\n%#v", rsp)
	fmt.Printf("%s\n", j)
}

func (h *handlers) deleteQueue(id string) {
	_, err := h.c.DeleteQueue(ctx, &proto.DeleteQueueRequest{QueueId: id})
	kingpin.FatalIfError(err, "cannot delete queue")
}

func (h *handlers) newQueue(store string, limit int64, tags map[string]string) {
	req := &proto.NewQueueRequest{
		Store: proto.Queue_Store(proto.Queue_Store_value[store]),
		Limit: limit,
		Tags:  tagsFromMap(tags),
	}
	rsp, err := h.c.NewQueue(ctx, req)
	kingpin.FatalIfError(err, "cannot create new queue")
	j, err := marshaller.MarshalToString(rsp)
	kingpin.FatalIfError(err, "cannot marshal new queue to JSON:\n%#v", rsp)
	fmt.Printf("%s\n", j)
}

func (h *handlers) addQueueTag(id, k, v string) {
	tag := &proto.Tag{Key: k, Value: v}
	_, err := h.c.AddQueueTag(ctx, &proto.AddQueueTagRequest{QueueId: id, Tag: tag})
	kingpin.FatalIfError(err, "cannot tag queue")
}

func (h *handlers) deleteQueueTag(id, k, v string) {
	tag := &proto.Tag{Key: k, Value: v}
	_, err := h.c.DeleteQueueTag(ctx, &proto.DeleteQueueTagRequest{QueueId: id, Tag: tag})
	kingpin.FatalIfError(err, "cannot untag queue")
}

func (h *handlers) addMessage(id string, tags map[string]string) {
	payload, err := ioutil.ReadAll(os.Stdin)
	kingpin.FatalIfError(err, "cannot read message payload from stdin")
	req := &proto.AddRequest{
		QueueId: id,
		Message: &proto.NewMessage{Payload: payload, Tags: tagsFromMap(tags)},
	}
	rsp, err := h.c.Add(ctx, req)
	kingpin.FatalIfError(err, "cannot add message to queue")
	j, err := marshaller.MarshalToString(rsp)
	kingpin.FatalIfError(err, "cannot marshal new message to JSON:\n%#v", rsp)
	fmt.Printf("%s\n", j)
}

func (h *handlers) popMessage(id string) {
	rsp, err := h.c.Pop(ctx, &proto.PopRequest{QueueId: id})
	kingpin.FatalIfError(err, "cannot pop message from queue")
	j, err := marshaller.MarshalToString(rsp)
	kingpin.FatalIfError(err, "cannot marshal popped message to JSON:\n%#v", rsp)
	fmt.Printf("%s\n", j)
}

func (h *handlers) peekMessage(id string) {
	rsp, err := h.c.Peek(ctx, &proto.PeekRequest{QueueId: id})
	kingpin.FatalIfError(err, "cannot peek at message in queue")
	j, err := marshaller.MarshalToString(rsp)
	kingpin.FatalIfError(err, "cannot marshal message to JSON:\n%#v", rsp)
	fmt.Printf("%s\n", j)
}

func tagsFromMap(tags map[string]string) []*proto.Tag {
	t := make([]*proto.Tag, 0, len(tags))
	for k, v := range tags {
		t = append(t, &proto.Tag{Key: k, Value: v})
	}
	return t
}
