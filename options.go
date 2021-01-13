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
