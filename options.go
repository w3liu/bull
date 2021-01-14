package bull

import (
	"context"
	"github.com/w3liu/bull/client"
	"github.com/w3liu/bull/registry"
	"github.com/w3liu/bull/server"
)

type Options struct {
	Client   client.Client
	Server   server.Server
	Registry registry.Registry

	// Before and After funcs
	BeforeStart []func() error
	BeforeStop  []func() error
	AfterStart  []func() error
	AfterStop   []func() error

	// Other options for implementations of the interface
	// can be stored in a context
	Context context.Context
}

func newOptions(opts ...Option) Options {
	opt := Options{
		Client:   client.DefaultClient,
		Server:   server.DefaultServer,
		Registry: registry.DefaultRegistry,
		Context:  context.Background(),
	}

	for _, o := range opts {
		o(&opt)
	}

	return opt
}

// Client to be used for service
func Client(c client.Client) Option {
	return func(o *Options) {
		o.Client = c
	}
}

// Context specifies a context for the service.
// Can be used to signal shutdown of the service and for extra option values.
func Context(ctx context.Context) Option {
	return func(o *Options) {
		o.Context = ctx
	}
}

// Server to be used for service
func Server(s server.Server) Option {
	return func(o *Options) {
		o.Server = s
	}
}

// Registry sets the registry for the service
// and the underlying components
func Registry(r registry.Registry) Option {
	return func(o *Options) {
		o.Registry = r
		// Update Client and Server
		o.Client.Init(client.Registry(r))
		o.Server.Init(server.Registry(r))
	}
}
