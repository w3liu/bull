package handler

import (
	"context"
	"github.com/w3liu/bull/client"
	proto "github.com/w3liu/bull/debug/proto"
)

// NewHandler returns an instance of the Debug Handler
func NewHandler(c client.Client) *Debug {
	return &Debug{}
}

type Debug struct {
}

func (d *Debug) Health(ctx context.Context, req *proto.HealthRequest, rsp *proto.HealthResponse) error {
	rsp.Status = "ok"
	return nil
}
