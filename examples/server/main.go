package main

import (
	"github.com/w3liu/bull"
	"github.com/w3liu/bull/examples/handler"
	pb "github.com/w3liu/bull/examples/proto"
	"github.com/w3liu/bull/registry"
	"google.golang.org/grpc"
)

func main() {
	r := registry.NewRegistry(registry.Addrs([]string{"127.0.0.1:2379"}...))
	service := bull.NewService(
		bull.Registry(r),
	)
	server := service.Server()
	grpcServer, ok := server.Instance().(*grpc.Server)
	if !ok {
		panic("not grpc server")
	}
	pb.RegisterPersonServer(grpcServer, &handler.Person{Name: "Foo"})
	err := service.Run()
	if err != nil {
		panic(err)
	}
}
