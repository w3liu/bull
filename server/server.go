package server

import (
	"github.com/google/uuid"
	"time"
)

var (
	DefaultAddress                 = ":0"
	DefaultName                    = "go.bull.server"
	DefaultVersion                 = "latest"
	DefaultId                      = uuid.New().String()
	DefaultServer           Server = newServer()
	DefaultRegisterInterval        = time.Second * 10
	DefaultRegisterTTL             = time.Second * 30
)

type Server interface {
	// Initialise options
	Init(...Option) error
	// Retrieve the options
	Options() Options
	// Start the server
	Start() error
	// Stop the server
	Stop() error
	// Server implementation
	String() string
	// Instance
	Instance() interface{}
}

type Request interface {
	// Service name requested
	Service() string
	// The action requested
	Method() string
	// Endpoint name requested
	Endpoint() string
	// Content type provided
	ContentType() string
	// Header of the request
	Header() map[string]string
	// Body is the initial decoded value
	Body() interface{}
	// Read the undecoded request body
	Read() ([]byte, error)
}

type Option func(*Options)

func NewServer(opts ...Option) Server {
	return newServer(opts...)
}
