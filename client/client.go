package client

import "time"

var (
	DefaultClient                             = newClient()
	DefaultDialTimeout                        = time.Second * 5
	NewClient          func(...Option) Client = newClient
)

type Client interface {
	Init(...Option) error
	Options() Options
	String() string
	Instance() interface{}
}

// Option used by the Client
type Option func(*Options)
