package registry

import "errors"

var (
	ErrNotFound = errors.New("service not found")
)

type Registry interface {
	Init(...Option) error
	Options() Options
	Register() error
}

type Service struct {
	Name      string            `json:"name"`
	Version   string            `json:"version"`
	Metadata  map[string]string `json:"metadata"`
	Endpoints []*Endpoint       `json:"endpoints"`
	Nodes     []*Node           `json:"nodes"`
}

type Node struct {
	Id       string            `json:"id"`
	Address  string            `json:"address"`
	Metadata map[string]string `json:"metadata"`
}

type Endpoint struct {
	Name     string            `json:"name"`
	Request  *Value            `json:"request"`
	Response *Value            `json:"response"`
	Metadata map[string]string `json:"metadata"`
}

type Value struct {
	Name   string   `json:"name"`
	Type   string   `json:"type"`
	Values []*Value `json:"values"`
}

type Option func(*Options)

type RegisterOption func(*RegisterOptions)

type WatchOption func(*WatchOptions)

type DeregisterOption func(*DeregisterOptions)

type GetOption func(*GetOptions)

type ListOption func(*ListOptions)
