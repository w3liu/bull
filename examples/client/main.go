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
	scheme := fmt.Sprintf("%s", registry.DefaultScheme)
	target := fmt.Sprintf("%s:///", scheme)
	r := registry.NewRegistry(registry.Addrs([]string{"127.0.0.1:2379"}...))
	registry.RegisterResolver(r, registry.ResolverScheme(scheme))

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cancel()

	conn, err := grpc.DialContext(ctx, target, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingPolicy": "%s"}`, roundrobin.Name)))

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
