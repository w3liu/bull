package client

import (
	"google.golang.org/grpc"
	"sync/atomic"
)

type grpcClient struct {
	once atomic.Value
	conn *grpc.ClientConn
	opts Options
}

func newClient(opts ...Option) Client {
	options := NewOptions()

	for _, o := range opts {
		o(&options)
	}

	rc := &grpcClient{
		opts: options,
	}

	rc.once.Store(false)

	c := Client(rc)

	return c
}

func (g *grpcClient) Init(opts ...Option) error {

	for _, o := range opts {
		o(&g.opts)
	}

	return nil
}

func (g *grpcClient) Options() Options {
	return g.opts
}

func (g *grpcClient) String() string {
	return "grpc"
}

func (g *grpcClient) Instance() interface{} {
	return g
}

func (g *grpcClient) getConn() *grpc.ClientConn {
	return g.conn
}
