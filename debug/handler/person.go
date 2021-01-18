package handler

import (
	"context"
	"fmt"
	pb "github.com/w3liu/bull/debug/proto/person"
)

type Person struct {
}

func (srv *Person) SayHello(ctx context.Context, in *pb.SayHelloRequest) (*pb.SayHelloResponse, error) {
	out := &pb.SayHelloResponse{
		Msg: fmt.Sprintf("hello %s", in.Name),
	}
	return out, nil
}
