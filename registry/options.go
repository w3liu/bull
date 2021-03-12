package registry

import (
	"context"
	"time"
)

type Options struct {
	Addrs   []string
	Timeout time.Duration
	Context context.Context
}

type RegisterOptions struct {
	TTL     time.Duration
	Context context.Context
}

type WatchOptions struct {
	// Specify a service to watch
	// If blank, the watch is for all services
	Service string
	// Other options for implementations of the interface
	// can be stored in a context
	Context context.Context
}

type DeregisterOptions struct {
	Context context.Context
}

type GetOptions struct {
	Context context.Context
}

type ListOptions struct {
	Context context.Context
}

type ResolverOptions struct {
	Scheme  string
	Service string
	TimeOut time.Duration
}

func Addrs(addrs ...string) Option {
	return func(o *Options) {
		o.Addrs = addrs
	}
}

func Timeout(t time.Duration) Option {
	return func(o *Options) {
		o.Timeout = t
	}
}

func RegisterTTL(t time.Duration) RegisterOption {
	return func(o *RegisterOptions) {
		o.TTL = t
	}
}

func RegisterContext(ctx context.Context) RegisterOption {
	return func(o *RegisterOptions) {
		o.Context = ctx
	}
}

func DeregisterContext(ctx context.Context) DeregisterOption {
	return func(o *DeregisterOptions) {
		o.Context = ctx
	}
}

func GetContext(ctx context.Context) GetOption {
	return func(o *GetOptions) {
		o.Context = ctx
	}
}

func ListContext(ctx context.Context) ListOption {
	return func(o *ListOptions) {
		o.Context = ctx
	}
}

func ResolverScheme(scheme string) ResolverOption {
	return func(o *ResolverOptions) {
		o.Scheme = scheme
	}
}

func ResolverService(service string) ResolverOption {
	return func(o *ResolverOptions) {
		o.Service = service
	}
}

// Watch a service
func WatchService(name string) WatchOption {
	return func(o *WatchOptions) {
		o.Service = name
	}
}
