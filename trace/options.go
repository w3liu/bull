package trace

type Options struct {
	Size int
}

type Option func(o *Options)

type ReadOptions struct {
	Trace string
}

type ReadOption func(o *ReadOptions)

func ReadTrace(t string) ReadOption {
	return func(o *ReadOptions) {
		o.Trace = t
	}
}

const (
	DefaultSize = 64
)

func DefaultOptions() Options {
	return Options{
		Size: DefaultSize,
	}
}
