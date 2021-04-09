package web

import "net/http"

type Service interface {
	// The service name
	Name() string
	// Init initialises options
	Init(...Option)
	// Options returns the current options
	Options() Options
	// Handle http request
	Handle(pattern string, handler http.Handler)
	// Run the service
	Run() error
	// The service implementation
	String() string
}

type Option func(*Options)

func NewService(opts ...Option) Service {
	return newService(opts...)
}
