// +build integration

package integration

import (
	"context"
	"net"
	"reflect"
	"testing"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/negz/q"
	"github.com/negz/q/manager"
	"github.com/negz/q/metrics"
	"github.com/negz/q/rpc"
	"github.com/negz/q/rpc/proto"
)

var ctx = context.Background()

func localhostWithRandomPort() (string, error) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", err
	}
	defer l.Close()
	return l.Addr().String(), nil
}

// TODO(negz): Make this table driven. Add tests for:
// * Tagging queues
// * Deleting queues
// * Adding to full queues
// * Popping from empty queues
func TestIntegration(t *testing.T) {
	mx, _ := metrics.NewPrometheus()
	m := manager.Instrumented(
		manager.New(),
		manager.WithMetrics(mx),
		manager.WithLogger(zap.NewNop()),
	)

	// This is a little racey. Something could consume the port immediately
	// between us closing it and attempting to reuse it.
	listen, err := localhostWithRandomPort()
	if err != nil {
		t.Fatal("Cannot find available port to listen on.")
	}
	l, err := net.Listen("tcp", listen)
	if err != nil {
		t.Fatalf("cannot listen on %v: %v", listen, err)
	}
	s := rpc.NewServer(l, m)
	go s.Serve()

	conn, err := grpc.Dial(listen, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("cannot dial server at %v: %v", listen, err)
	}
	defer conn.Close()
	client := proto.NewQClient(conn)

	id, err := newQueue(client)
	if err != nil {
		t.Fatalf("cannot create new queue: %v", err)
	}

	payload := []byte("dove")
	if err := newMessage(client, id, payload); err != nil {
		t.Fatalf("cannot create new message: %v", err)
	}

	peek, err := peekMessage(client, id)
	if err != nil {
		t.Fatalf("cannot peek into queue: %v", err)
	}
	if !reflect.DeepEqual(peek, payload) {
		t.Fatalf("peek: want %s, got %s", payload, peek)
	}

	pop, err := popMessage(client, id)
	if err != nil {
		t.Fatalf("cannot pop from queue: %v", err)
	}
	if !reflect.DeepEqual(pop, payload) {
		t.Fatalf("pop: want %s, got %s", payload, pop)
	}
}

func newQueue(c proto.QClient) (string, error) {
	req := &proto.NewQueueRequest{
		Store: proto.MEMORY,
		Limit: int64(q.Unbounded),
		Tags:  []*proto.Tag{&proto.Tag{Key: "type", Value: "cubesat launcher"}},
	}
	rsp, err := c.NewQueue(ctx, req)
	if err != nil {
		return "", err
	}
	return rsp.GetQueue().GetMeta().GetId(), nil
}

func newMessage(c proto.QClient, id string, payload []byte) error {
	req := &proto.AddRequest{
		QueueId: id,
		Message: &proto.NewMessage{
			Payload: payload,
			Tags:    []*proto.Tag{&proto.Tag{Key: "type", Value: "satellite"}},
		},
	}
	_, err := c.Add(ctx, req)
	return err
}

func peekMessage(c proto.QClient, id string) ([]byte, error) {
	rsp, err := c.Peek(ctx, &proto.PeekRequest{QueueId: id})
	if err != nil {
		return nil, err
	}
	return rsp.GetMessage().GetPayload(), nil
}

func popMessage(c proto.QClient, id string) ([]byte, error) {
	rsp, err := c.Pop(ctx, &proto.PopRequest{QueueId: id})
	if err != nil {
		return nil, err
	}
	return rsp.GetMessage().GetPayload(), nil
}
