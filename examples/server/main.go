package main

import (
	"context"
	"fmt"
	"github.com/w3liu/bull"
	pb "github.com/w3liu/bull/examples/proto"
	"github.com/w3liu/bull/registry"
	"google.golang.org/grpc"
)

func main() {
	r := registry.NewRegistry(registry.Addrs([]string{"192.168.10.20:2379"}...))
	service := bull.NewService(
		bull.Registry(r),
	)
	server := service.Server()
	grpcServer := server.Instance().(*grpc.Server)
	pb.RegisterPersonServer(grpcServer, &person{})
	err := service.Run()
	if err != nil {
		panic(err)
	}
}

type person struct {
}

func (srv *person) SayHello(ctx context.Context, in *pb.SayHelloRequest) (*pb.SayHelloResponse, error) {
	out := &pb.SayHelloResponse{
		Msg: fmt.Sprintf("hello %s 2", in.Name),
	}
	return out, nil
}
