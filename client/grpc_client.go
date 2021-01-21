package client

import (
	"github.com/pkg/errors"
	"github.com/w3liu/bull/selector"
	"sync/atomic"
)

type grpcClient struct {
	once atomic.Value
	pool *pool
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

	rc.pool = newPool(options.PoolSize, options.PoolTTL, rc.poolMaxIdle(), rc.poolMaxStreams())

	c := Client(rc)

	return c
}

func (g *grpcClient) poolMaxStreams() int {
	if g.opts.Context == nil {
		return DefaultPoolMaxStreams
	}
	v := g.opts.Context.Value(poolMaxStreams{})
	if v == nil {
		return DefaultPoolMaxStreams
	}
	return v.(int)
}

func (g *grpcClient) poolMaxIdle() int {
	if g.opts.Context == nil {
		return DefaultPoolMaxIdle
	}
	v := g.opts.Context.Value(poolMaxIdle{})
	if v == nil {
		return DefaultPoolMaxIdle
	}
	return v.(int)
}

func (g *grpcClient) next(name string, opts CallOptions) (selector.Next, error) {
	service := name
	// get next nodes from the selector
	next, err := g.opts.Selector.Select(service, opts.SelectOptions...)
	if err != nil {
		if err == selector.ErrNotFound {
			return nil, errors.Errorf("go.micro.client", "service %s: %s", service, err.Error())
		}
		return nil, errors.Errorf("go.micro.client", "error selecting %s node: %s", service, err.Error())
	}

	return next, nil
}

func (g *grpcClient) Init(opts ...Option) error {
	size := g.opts.PoolSize
	ttl := g.opts.PoolTTL

	for _, o := range opts {
		o(&g.opts)
	}

	// update pool configuration if the options changed
	if size != g.opts.PoolSize || ttl != g.opts.PoolTTL {
		g.pool.Lock()
		g.pool.size = g.opts.PoolSize
		g.pool.ttl = int64(g.opts.PoolTTL.Seconds())
		g.pool.Unlock()
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
