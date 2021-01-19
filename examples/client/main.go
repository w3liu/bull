package main

import (
	"context"
	"fmt"
	pb "github.com/w3liu/bull/examples/proto"
	"github.com/w3liu/bull/registry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"time"
)

func main() {
	r := registry.NewRegistry(registry.Addrs([]string{"192.168.10.20:2379"}...))
	registry.RegisterResolver(r)

	conn, err := grpc.Dial(fmt.Sprintf("%s:///", registry.DefaultScheme), grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, roundrobin.Name)))

	if err != nil {
		panic(err)
	}

	client := pb.NewPersonClient(conn)

	for i := 0; i < 10; i++ {
		ctx, _ := context.WithTimeout(context.TODO(), time.Second*5)
		rsp, err := client.SayHello(ctx, &pb.SayHelloRequest{
			Name: "Foo",
		})
		if err != nil {
			panic(err)
		}
		fmt.Println(rsp)
	}
}
