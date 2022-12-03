package main

import (
	"github.com/w3liu/bull"
	"github.com/w3liu/bull/examples/handler"
	pb "github.com/w3liu/bull/examples/proto"
	"github.com/w3liu/bull/registry"
	"github.com/w3liu/bull/server"
)

func main() {
	r := registry.NewRegistry(registry.Addrs([]string{"127.0.0.1:2379"}...))

	srv := server.NewGrpcServer(server.Name("hello.svc"))

	service := bull.NewService(
		bull.Registry(r),
		bull.Server(srv),
	)

	pb.RegisterPersonServer(srv.Server, &handler.Person{Name: "Foo"})
	err := service.Run()
	if err != nil {
		panic(err)
	}
}
