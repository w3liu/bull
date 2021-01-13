package server

import "context"

type HandlerFunc func(ctx context.Context, req Request) (rsp interface{}, err error)

// HandlerWrapper wraps the HandlerFunc and returns the equivalent
type HandlerWrapper func(HandlerFunc) HandlerFunc
