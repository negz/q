package rpc

import (
	"net"

	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/negz/q"
	"github.com/negz/q/e"
	"github.com/negz/q/memory"
	"github.com/negz/q/rpc/proto"
)

type qServer struct {
	m q.Manager
}

func newQServer(m q.Manager) proto.QServer {
	return &qServer{m}
}

func (s *qServer) ListQueues(_ context.Context, _ *proto.ListQueuesRequest) (*proto.ListQueuesResponse, error) {
	l, err := s.m.List()
	if err != nil {
		return nil, e.GRPC(err)
	}
	queues := make([]*proto.Queue, 0, len(l))
	for _, queue := range l {
		queues = append(queues, queueToProto(queue))
	}
	return &proto.ListQueuesResponse{Queues: queues}, nil
}

func (s *qServer) NewQueue(_ context.Context, r *proto.NewQueueRequest) (*proto.NewQueueResponse, error) {
	// TODO(negz): Implement a queue factory if/when we have more than one kind
	// of backing store. :)
	tags := tagsFromProto(r.GetTags())
	queue := memory.New(memory.Limit(int(r.GetLimit())), memory.Tagged(tags...))
	return &proto.NewQueueResponse{Queue: queueToProto(queue)}, nil
}

func (s *qServer) GetQueue(_ context.Context, r *proto.GetQueueRequest) (*proto.GetQueueResponse, error) {
	id, err := parseID(r.GetQueueId())
	if err != nil {
		return nil, e.GRPC(err)
	}
	queue, err := s.m.Get(id)
	if err != nil {
		return nil, e.GRPC(err)
	}
	return &proto.GetQueueResponse{Queue: queueToProto(queue)}, nil
}

func (s *qServer) DeleteQueue(_ context.Context, r *proto.DeleteQueueRequest) (*proto.DeleteQueueResponse, error) {
	id, err := parseID(r.GetQueueId())
	if err != nil {
		return nil, e.GRPC(err)
	}
	if err := s.m.Delete(id); err != nil {
		return nil, e.GRPC(err)
	}
	return &proto.DeleteQueueResponse{}, nil
}

func (s *qServer) AddQueueTag(_ context.Context, r *proto.AddQueueTagRequest) (*proto.AddQueueTagResponse, error) {
	id, err := parseID(r.GetQueueId())
	if err != nil {
		return nil, e.GRPC(err)
	}
	queue, err := s.m.Get(id)
	if err != nil {
		return nil, e.GRPC(err)
	}
	tag := r.GetTag()
	if tag == nil {
		return nil, e.GRPC(e.ErrInvalid(errors.New("did not supply a tag to add")))
	}
	queue.Tags().Add(tag.Key, tag.Value)
	return &proto.AddQueueTagResponse{}, nil
}

func (s *qServer) DeleteQueueTag(_ context.Context, r *proto.DeleteQueueTagRequest) (*proto.DeleteQueueTagResponse, error) {
	id, err := parseID(r.GetQueueId())
	if err != nil {
		return nil, e.GRPC(err)
	}
	queue, err := s.m.Get(id)
	if err != nil {
		return nil, e.GRPC(err)
	}
	tag := r.GetTag()
	if tag == nil {
		return nil, e.GRPC(e.ErrInvalid(errors.New("did not supply a tag to delete")))
	}
	queue.Tags().Remove(tag.Key, tag.Value)
	return &proto.DeleteQueueTagResponse{}, nil
}

func (s *qServer) Add(_ context.Context, r *proto.AddRequest) (*proto.AddResponse, error) {
	id, err := parseID(r.GetQueueId())
	if err != nil {
		return nil, e.GRPC(err)
	}
	queue, err := s.m.Get(id)
	if err != nil {
		return nil, e.GRPC(err)
	}
	p := r.GetMessage().GetPayload()
	tags := tagsFromProto(r.GetMessage().GetMeta().GetTags())
	if err := queue.Add(q.NewMessage(p, q.Tagged(tags...))); err != nil {
		return nil, e.GRPC(err)
	}
	return &proto.AddResponse{}, nil
}

func (s *qServer) Pop(_ context.Context, r *proto.PopRequest) (*proto.PopResponse, error) {
	id, err := parseID(r.GetQueueId())
	if err != nil {
		return nil, e.GRPC(err)
	}
	queue, err := s.m.Get(id)
	if err != nil {
		return nil, e.GRPC(err)
	}
	m, err := queue.Pop()
	if err != nil {
		return nil, e.GRPC(err)
	}
	return &proto.PopResponse{Message: messageToProto(m)}, nil
}

func (s *qServer) Peek(_ context.Context, r *proto.PeekRequest) (*proto.PeekResponse, error) {
	id, err := parseID(r.GetQueueId())
	if err != nil {
		return nil, e.GRPC(err)
	}
	queue, err := s.m.Get(id)
	if err != nil {
		return nil, e.GRPC(err)
	}
	m, err := queue.Peek()
	if err != nil {
		return nil, e.GRPC(err)
	}
	return &proto.PeekResponse{Message: messageToProto(m)}, nil
}

// A Server serves gRPC requests.
type Server struct {
	l net.Listener
	m q.Manager
}

// NewServer returns a new gRPC server.
func NewServer(l net.Listener, m q.Manager) *Server {
	return &Server{l: l, m: m}
}

// Serve begins serving gRPC requests.
func (s *Server) Serve() error {
	g := grpc.NewServer()
	proto.RegisterQServer(g, newQServer(s.m))
	return errors.Wrap(g.Serve(s.l), "cannot serve gRPC requests")
}
