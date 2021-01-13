package client

type Client interface {
	Init(...Option) error
	Options() Options
	String() string
}

// Option used by the Client
type Option func(*Options)
