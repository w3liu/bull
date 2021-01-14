package client

import (
	"context"
	"github.com/w3liu/bull/registry"
	"github.com/w3liu/bull/selector"
	"time"
)

type Options struct {
	Registry registry.Registry
	Selector selector.Selector

	PoolSize int
	PoolTTL  time.Duration

	// Default Call Options
	CallOptions CallOptions

	Context context.Context
}

type CallOptions struct {
	SelectOptions []selector.SelectOption

	// Address of remote hosts
	Address []string
	// Transport Dial Timeout
	DialTimeout time.Duration
	// Request/Response timeout
	RequestTimeout time.Duration

	// Other options for implementations of the interface
	// can be stored in a context
	Context context.Context
}

func NewOptions(options ...Option) Options {
	opts := Options{
		Registry: nil,
		Selector: selector.DefaultSelector,
		CallOptions: CallOptions{
			SelectOptions:  nil,
			Address:        nil,
			DialTimeout:    0,
			RequestTimeout: 0,
			Context:        nil,
		},
		Context: context.Background(),
	}

	for _, o := range options {
		o(&opts)
	}

	return opts
}

// Registry to find nodes for a given service
func Registry(r registry.Registry) Option {
	return func(o *Options) {
		o.Registry = r
		// set in the selector
		o.Selector.Init(selector.Registry(r))
	}
}

// Select is used to select a node to route a request to
func Selector(s selector.Selector) Option {
	return func(o *Options) {
		o.Selector = s
	}
}
