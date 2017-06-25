// +build integration

package integration

import (
	"context"
	"net"
	"reflect"
	"testing"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/negz/q/manager"
	"github.com/negz/q/metrics"
	"github.com/negz/q/rpc"
	"github.com/negz/q/rpc/proto"
)

const Unbounded int64 = -1

var ctx = context.Background()

type message struct {
	payload []byte
	tags    []*proto.Tag
}

var integrationTests = []struct {
	store    proto.Queue_Store
	limit    int64
	tags     []*proto.Tag
	messages []*message
}{
	{
		store: proto.MEMORY,
		limit: Unbounded,
		tags:  []*proto.Tag{&proto.Tag{"type", "cubesat launcher"}},
		messages: []*message{
			&message{
				payload: []byte("dove 001"),
				tags:    []*proto.Tag{&proto.Tag{"size", "3U"}},
			},
			&message{
				payload: []byte("dove 002"),
				tags:    []*proto.Tag{&proto.Tag{"size", "3U"}},
			},
		},
	},
	{
		store: proto.MEMORY,
		limit: 1,
		tags:  []*proto.Tag{&proto.Tag{"type", "cubesat launcher"}},
		messages: []*message{
			&message{
				payload: []byte("dove 001"),
				tags:    []*proto.Tag{&proto.Tag{"size", "3U"}},
			},
			&message{
				payload: []byte("dove 002"),
				tags:    []*proto.Tag{&proto.Tag{"size", "3U"}},
			},
		},
	},
}

func TestIntegration(t *testing.T) {
	// This is a little racey. Something could consume the port immediately
	// between us closing it and attempting to reuse it.
	listen, err := localhostWithRandomPort()
	if err != nil {
		t.Fatal("Cannot find available port to listen on.")
	}
	conn, err := newServer(listen)
	if err != nil {
		t.Fatal("Cannot create new server: %v", err)
	}
	defer conn.Close()
	c := &itClient{proto.NewQClient(conn)}

	for _, tt := range integrationTests {
		id, err := c.newQueue(tt.limit, tt.store, tt.tags...)
		if err != nil {
			t.Errorf("c.newQueue(%v, %v, %v): %v", tt.limit, tt.store, tt.tags, err)
		}

		t.Run("AddAll", func(t *testing.T) {
			for i, m := range tt.messages {
				if err := c.newMessage(id, m.payload, m.tags...); err != nil {
					// We expect an error if we're trying to add more messages than
					// the queue can fit.
					if int64(i) >= tt.limit-1 {
						s, ok := status.FromError(err)
						if ok && s.Code() == codes.ResourceExhausted {
							continue
						}
					}
					t.Errorf("c.newMessage(%v, %v, %v): %v", id, m.payload, m.tags, err)
				}
			}
		})

		t.Run("PeekFirst", func(t *testing.T) {
			if len(tt.messages) < 1 {
				return
			}
			payload, err := c.peekMessage(id)
			if err != nil {
				t.Errorf("c.peekMessage(%v): %v", id, err)
			}
			if !reflect.DeepEqual(payload, tt.messages[0].payload) {
				t.Errorf("c.popMessage(%v): want %s, got %s", tt.messages[0].payload, payload)
			}
		})

		t.Run("PopAll", func(t *testing.T) {
			for i, m := range tt.messages {
				payload, err := c.popMessage(id)
				if err != nil {
					// We expect an error if we're trying to pop more messages than the
					// queue could fit.
					if int64(i) >= tt.limit-1 {
						s, ok := status.FromError(err)
						if ok && s.Code() == codes.NotFound {
							continue
						}
					}
					t.Errorf("c.popMessage(%v): %v", id, err)
					continue
				}
				if !reflect.DeepEqual(payload, m.payload) {
					t.Errorf("c.popMessage(%v): want %s, got %s", m.payload, payload)
				}
			}
		})

		t.Run("PeekEmpty", func(t *testing.T) {
			_, err := c.peekMessage(id)
			s, ok := status.FromError(err)
			if !ok || s.Code() != codes.NotFound {
				t.Errorf("c.peekMessage(%v): want not found error, got %v", id, err)
			}
		})

		t.Run("DeleteQueue", func(t *testing.T) {
			if err := c.deleteQueue(id); err != nil {
				t.Errorf("c.deleteQueue(%v): %v", id, err)
			}
		})
	}
}

func localhostWithRandomPort() (string, error) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", err
	}
	defer l.Close()
	return l.Addr().String(), nil
}

func newServer(listen string) (*grpc.ClientConn, error) {
	mx, _ := metrics.NewPrometheus()
	m := manager.Instrumented(
		manager.New(),
		manager.WithMetrics(mx),
		manager.WithLogger(zap.NewNop()),
	)
	l, err := net.Listen("tcp", listen)
	if err != nil {
		return nil, err
	}
	s := rpc.NewServer(l, m)
	go s.Serve()
	return grpc.Dial(listen, grpc.WithInsecure())
}

type itClient struct {
	c proto.QClient
}

func (c *itClient) newQueue(limit int64, s proto.Queue_Store, t ...*proto.Tag) (string, error) {
	req := &proto.NewQueueRequest{
		Store: s,
		Limit: limit,
		Tags:  t,
	}
	rsp, err := c.c.NewQueue(ctx, req)
	if err != nil {
		return "", err
	}
	return rsp.GetQueue().GetMeta().GetId(), nil
}

func (c *itClient) deleteQueue(id string) error {
	_, err := c.c.DeleteQueue(ctx, &proto.DeleteQueueRequest{QueueId: id})
	return err
}

func (c *itClient) newMessage(id string, payload []byte, t ...*proto.Tag) error {
	req := &proto.AddRequest{
		QueueId: id,
		Message: &proto.NewMessage{Payload: payload, Tags: t},
	}
	_, err := c.c.Add(ctx, req)
	return err
}

func (c *itClient) peekMessage(id string) ([]byte, error) {
	rsp, err := c.c.Peek(ctx, &proto.PeekRequest{QueueId: id})
	if err != nil {
		return nil, err
	}
	return rsp.GetMessage().GetPayload(), nil
}

func (c *itClient) popMessage(id string) ([]byte, error) {
	rsp, err := c.c.Pop(ctx, &proto.PopRequest{QueueId: id})
	if err != nil {
		return nil, err
	}
	return rsp.GetMessage().GetPayload(), nil
}
