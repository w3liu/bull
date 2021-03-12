package main

import (
	"context"
	"fmt"
	"github.com/w3liu/bull/client"
	pb "github.com/w3liu/bull/examples/proto"
	"github.com/w3liu/bull/registry"
	"google.golang.org/grpc"
	"time"
)

func main() {
	r := registry.NewRegistry(registry.Addrs([]string{"127.0.0.1:2379"}...))
	cli := client.NewClient(
		client.Registry(r),
		client.Service("hello.svc"))

	conn, ok := cli.Instance().(*grpc.ClientConn)
	if !ok {
		panic("not grpc client conn instance")
	}
	personClient := pb.NewPersonClient(conn)

	for i := 0; i < 10; i++ {
		ctx, _ := context.WithTimeout(context.TODO(), time.Second*5)
		rsp, err := personClient.SayHello(ctx, &pb.SayHelloRequest{
			Name: "Bar",
		})
		if err != nil {
			panic(err)
		}
		fmt.Println(rsp)
	}
}
