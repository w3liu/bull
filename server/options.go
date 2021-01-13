package server

import (
	"github.com/w3liu/bull/registry"
	"time"
)

type Options struct {
	Registry     registry.Registry
	Name         string
	Address      string
	Id           string
	Version      string
	HdlrWrappers []HandlerWrapper

	RegisterTTL      time.Duration
	RegisterInterval time.Duration
}
