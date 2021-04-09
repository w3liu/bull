package web

import (
	"context"
	"github.com/w3liu/bull/registry"
)

var (
	DefaultName    = "go.bull.web"
	DefaultAddress = ":8090"
)

type Options struct {
	Registry registry.Registry
	Context  context.Context
	Name     string
	Address  string
}

func newOptions(opts ...Option) Options {
	options := Options{
		Name:     DefaultName,
		Address:  DefaultAddress,
		Registry: registry.DefaultRegistry,
		Context:  context.Background(),
	}
	for _, o := range opts {
		o(&options)
	}
	return options
}

func Name(name string) Option {
	return func(options *Options) {
		options.Name = name
	}
}

func Context(ctx context.Context) Option {
	return func(options *Options) {
		options.Context = ctx
	}
}

func Address(addr string) Option {
	return func(options *Options) {
		options.Address = addr
	}
}
