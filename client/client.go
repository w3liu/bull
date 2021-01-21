package client

import "time"

var (
	DefaultClient         = newClient()
	DefaultRequestTimeout = time.Second * 5
	// DefaultPoolSize sets the connection pool size
	DefaultPoolSize = 100
	// DefaultPoolTTL sets the connection pool ttl
	DefaultPoolTTL                        = time.Minute
	NewClient      func(...Option) Client = newClient
)

type Client interface {
	Init(...Option) error
	Options() Options
	String() string
	Instance() interface{}
}

// Option used by the Client
type Option func(*Options)

type CallOption func(*CallOptions)
