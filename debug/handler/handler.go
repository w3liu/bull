package handler

import (
	"context"
	"fmt"
	proto "github.com/w3liu/bull/debug/proto"
	person "github.com/w3liu/bull/debug/proto/person"
)

// NewHandler returns an instance of the Debug Handler
func NewHandler() *Debug {
	return &Debug{}
}

type Debug struct {
}

func (d *Debug) Health(ctx context.Context, req *proto.HealthRequest, rsp *proto.HealthResponse) error {
	rsp.Status = "ok"
	return nil
}

type PersonService struct {
	Name string
}

func NewPersonService(name string) *PersonService {
	return &PersonService{Name: name}
}

func (s *PersonService) SayHello(ctx context.Context, in *person.SayHelloRequest, out *person.SayHelloResponse) error {
	out = &person.SayHelloResponse{
		Msg: fmt.Sprintf("hello %s, I am %s", in.Name, s.Name),
	}
	return nil
}
