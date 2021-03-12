package client

import (
	"github.com/w3liu/bull/registry"
	"time"
)

type Options struct {
	Registry    registry.Registry
	Service     string
	DialTimeout time.Duration
}

func NewOptions(options ...Option) Options {
	opts := Options{
		Registry:    nil,
		Service:     "",
		DialTimeout: DefaultDialTimeout,
	}

	for _, o := range options {
		o(&opts)
	}

	return opts
}

func Registry(r registry.Registry) Option {
	return func(o *Options) {
		o.Registry = r
	}
}

func Service(s string) Option {
	return func(o *Options) {
		o.Service = s
	}
}

func DialTimeout(t time.Duration) Option {
	return func(o *Options) {
		o.DialTimeout = t
	}
}
