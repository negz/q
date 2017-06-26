// Package rpc implements a gRPC server for the q queue service.
package rpc

import (
	"net"

	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/negz/q"
	"github.com/negz/q/e"
	"github.com/negz/q/factory"
	"github.com/negz/q/proto"
)

// A Server serves gRPC requests.
type Server struct {
	l net.Listener
	f q.Factory
	m q.Manager
}

// An Option represents an optional argument to a new server.
type Option func(*Server)

// WithQueueFactory specifies a bespoke queue factory to be used to create a
// server's new queues.
func WithQueueFactory(f q.Factory) Option {
	return func(s *Server) {
		s.f = f
	}
}

// NewServer returns a new gRPC server.
func NewServer(l net.Listener, m q.Manager, o ...Option) *Server {
	s := &Server{l: l, f: factory.Default, m: m}
	for _, opt := range o {
		opt(s)
	}
	return s
}

// Serve gRPC requests forever.
func (s *Server) Serve() error {
	g := grpc.NewServer()
	proto.RegisterQServer(g, &qServer{s.f, s.m})
	return errors.Wrap(g.Serve(s.l), "cannot serve gRPC requests")
}

type qServer struct {
	f q.Factory
	m q.Manager
}

func (s *qServer) ListQueues(_ context.Context, _ *proto.ListQueuesRequest) (*proto.ListQueuesResponse, error) {
	l, err := s.m.List()
	if err != nil {
		return nil, e.GRPC(errors.Wrap(err, "cannot list queues"))
	}
	queues := make([]*proto.Queue, 0, len(l))
	for _, queue := range l {
		pq, err := proto.FromQueue(queue)
		if err != nil {
			return nil, e.GRPC(errors.Wrap(err, "cannot marshal queue to protobuf"))
		}
		queues = append(queues, pq)
	}
	return &proto.ListQueuesResponse{Queues: queues}, nil
}

func (s *qServer) NewQueue(_ context.Context, r *proto.NewQueueRequest) (*proto.NewQueueResponse, error) {
	tags := proto.ToTags(r.GetTags())
	queue, err := s.f.New(proto.ToStore[r.GetStore()], int(r.GetLimit()), tags...)
	if err != nil {
		return nil, e.GRPC(errors.Wrap(err, "cannot create new queue"))
	}
	if aerr := s.m.Add(queue); aerr != nil {
		return nil, e.GRPC(errors.Wrap(aerr, "cannot add queue to manager"))
	}
	pq, err := proto.FromQueue(queue)
	if err != nil {
		return nil, e.GRPC(errors.Wrap(err, "cannot marshal queue to protobuf"))
	}
	return &proto.NewQueueResponse{Queue: pq}, nil
}

func (s *qServer) GetQueue(_ context.Context, r *proto.GetQueueRequest) (*proto.GetQueueResponse, error) {
	id, err := proto.ParseID(r.GetQueueId())
	if err != nil {
		return nil, e.GRPC(errors.Wrap(err, "cannot parse ID"))
	}
	queue, err := s.m.Get(id)
	if err != nil {
		return nil, e.GRPC(errors.Wrapf(err, "cannot get queue %s", id))
	}
	pq, err := proto.FromQueue(queue)
	if err != nil {
		return nil, e.GRPC(errors.Wrap(err, "cannot marshal queue to protobuf"))
	}
	return &proto.GetQueueResponse{Queue: pq}, nil
}

func (s *qServer) DeleteQueue(_ context.Context, r *proto.DeleteQueueRequest) (*proto.DeleteQueueResponse, error) {
	id, err := proto.ParseID(r.GetQueueId())
	if err != nil {
		return nil, e.GRPC(errors.Wrap(err, "cannot parse ID"))
	}
	if err := s.m.Delete(id); err != nil {
		return nil, e.GRPC(errors.Wrapf(err, "cannot delete queue %s", id))
	}
	return &proto.DeleteQueueResponse{}, nil
}

func (s *qServer) AddQueueTag(_ context.Context, r *proto.AddQueueTagRequest) (*proto.AddQueueTagResponse, error) {
	id, err := proto.ParseID(r.GetQueueId())
	if err != nil {
		return nil, e.GRPC(errors.Wrap(err, "cannot parse ID"))
	}
	queue, err := s.m.Get(id)
	if err != nil {
		return nil, e.GRPC(errors.Wrapf(err, "cannot get queue %s", id))
	}
	tag := r.GetTag()
	if tag == nil {
		return nil, e.GRPC(e.ErrInvalid(errors.New("did not supply a tag to add")))
	}
	queue.Tags().Add(tag.Key, tag.Value)
	return &proto.AddQueueTagResponse{}, nil
}

func (s *qServer) DeleteQueueTag(_ context.Context, r *proto.DeleteQueueTagRequest) (*proto.DeleteQueueTagResponse, error) {
	id, err := proto.ParseID(r.GetQueueId())
	if err != nil {
		return nil, e.GRPC(errors.Wrap(err, "cannot parse ID"))
	}
	queue, err := s.m.Get(id)
	if err != nil {
		return nil, e.GRPC(errors.Wrapf(err, "cannot get queue %s", id))
	}
	tag := r.GetTag()
	if tag == nil {
		return nil, e.GRPC(e.ErrInvalid(errors.New("did not supply a tag to delete")))
	}
	queue.Tags().Remove(tag.Key, tag.Value)
	return &proto.DeleteQueueTagResponse{}, nil
}

func (s *qServer) Add(_ context.Context, r *proto.AddRequest) (*proto.AddResponse, error) {
	id, err := proto.ParseID(r.GetQueueId())
	if err != nil {
		return nil, e.GRPC(errors.Wrap(err, "cannot parse ID"))
	}
	queue, err := s.m.Get(id)
	if err != nil {
		return nil, e.GRPC(errors.Wrapf(err, "cannot get queue %s", id))
	}
	p := r.GetMessage().GetPayload()
	tags := proto.ToTags(r.GetMessage().GetTags())
	m := q.NewMessage(p, q.Tagged(tags...))
	if aerr := queue.Add(m); aerr != nil {
		return nil, e.GRPC(errors.Wrap(aerr, "cannot add message to queue"))
	}
	pm, err := proto.FromMessage(m)
	if err != nil {
		return nil, e.GRPC(errors.Wrap(err, "cannot marshal message to protobuf"))
	}
	return &proto.AddResponse{Message: pm}, nil
}

func (s *qServer) Pop(_ context.Context, r *proto.PopRequest) (*proto.PopResponse, error) {
	id, err := proto.ParseID(r.GetQueueId())
	if err != nil {
		return nil, e.GRPC(errors.Wrap(err, "cannot parse ID"))
	}
	queue, err := s.m.Get(id)
	if err != nil {
		return nil, e.GRPC(errors.Wrapf(err, "cannot get queue %s", id))
	}
	m, err := queue.Pop()
	if err != nil {
		return nil, e.GRPC(errors.Wrap(err, "cannot pop message from queue"))
	}
	pm, err := proto.FromMessage(m)
	if err != nil {
		return nil, e.GRPC(errors.Wrap(err, "cannot marshal message to protobuf"))
	}
	return &proto.PopResponse{Message: pm}, nil
}

func (s *qServer) Peek(_ context.Context, r *proto.PeekRequest) (*proto.PeekResponse, error) {
	id, err := proto.ParseID(r.GetQueueId())
	if err != nil {
		return nil, e.GRPC(errors.Wrap(err, "cannot parse ID"))
	}
	queue, err := s.m.Get(id)
	if err != nil {
		return nil, e.GRPC(errors.Wrapf(err, "cannot get queue %s", id))
	}
	m, err := queue.Peek()
	if err != nil {
		return nil, e.GRPC(errors.Wrap(err, "cannot peek into queue"))
	}
	pm, err := proto.FromMessage(m)
	if err != nil {
		return nil, e.GRPC(errors.Wrap(err, "cannot marshal message to protobuf"))
	}
	return &proto.PeekResponse{Message: pm}, nil
}
