package client

import (
	"context"
	"fmt"
	"github.com/w3liu/bull/logger"
	"github.com/w3liu/bull/registry"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"sync"
)

type grpcClient struct {
	sync.Mutex
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
	g.Lock()
	defer g.Unlock()
	conn, err := g.getConn()
	if err != nil {
		logger.Error("g.getConn() error", zap.Error(err))
		return nil
	}
	return conn
}

func (g *grpcClient) getConn() (*grpc.ClientConn, error) {
	if g.conn != nil {
		return g.conn, nil
	}
	scheme := fmt.Sprintf("%s", registry.DefaultScheme)
	target := fmt.Sprintf("%s:///", scheme)
	resolverOptions := []registry.ResolverOption{
		registry.ResolverScheme(scheme),
	}
	if g.opts.Service != "" {
		registry.ResolverService(g.opts.Service)
	}
	registry.RegisterResolver(g.opts.Registry, resolverOptions...)

	ctx, cancel := context.WithTimeout(context.TODO(), g.opts.DialTimeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, target, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingPolicy": "%s"}`, roundrobin.Name)))
	if err != nil {
		return nil, err
	}
	g.conn = conn
	return g.conn, nil
}
