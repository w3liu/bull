package pool

import "time"

// Pool is an interface for connection pooling
type Pool interface {
	// Close the pool
	Close() error
	// Get a connection
	Get(addr string) (Conn, error)
	// Releaes the connection
	Release(c Conn, status error) error
}

type Conn interface {
	// unique id of connection
	Id() string
	// time it was created
	Created() time.Time
}
