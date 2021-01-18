package server

import (
	"context"
	"github.com/w3liu/bull/registry"
	"sync"
)

var (
	// DefaultMaxMsgSize define maximum message size that server can send
	// or receive.  Default value is 4MB.
	DefaultMaxMsgSize = 1024 * 1024 * 4
)

func newOptions(opt ...Option) Options {
	opts := Options{
		Name:             DefaultName,
		Address:          DefaultAddress,
		Id:               DefaultId,
		Version:          DefaultVersion,
		Metadata:         make(map[string]string),
		RegisterTTL:      DefaultRegisterTTL,
		RegisterInterval: DefaultRegisterInterval,
	}

	for _, o := range opt {
		o(&opts)
	}

	if opts.Registry == nil {
		opts.Registry = registry.DefaultRegistry
	}

	return opts
}

func wait(ctx context.Context) *sync.WaitGroup {
	if ctx == nil {
		return nil
	}
	wg, ok := ctx.Value("wait").(*sync.WaitGroup)
	if !ok {
		return nil
	}
	return wg
}
