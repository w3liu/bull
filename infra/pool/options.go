package pool

import "time"

type Options struct {
	TTL  time.Duration
	Size int
}

type Option func(*Options)

func Size(i int) Option {
	return func(o *Options) {
		o.Size = i
	}
}

func TTL(t time.Duration) Option {
	return func(o *Options) {
		o.TTL = t
	}
}
