package handler

import (
	"context"
	"fmt"
	pb "github.com/w3liu/bull/examples/proto"
)

type Person struct {
	Name string
}

func (srv *Person) SayHello(ctx context.Context, in *pb.SayHelloRequest) (*pb.SayHelloResponse, error) {
	out := &pb.SayHelloResponse{
		Msg: fmt.Sprintf("%s say hello to %s", srv.Name, in.Name),
	}
	return out, nil
}
