package bull

import (
	"context"
	"fmt"
	"github.com/w3liu/bull/debug/handler"
	"github.com/w3liu/bull/debug/proto/person"
	"github.com/w3liu/bull/registry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"testing"
	"time"
)

func TestService(t *testing.T) {
	r := registry.NewRegistry(registry.Addrs([]string{"192.168.10.20:2379"}...))
	service := NewService(
		Registry(r),
	)
	server := service.Server()
	grpcServer := server.Instance().(*grpc.Server)
	person.RegisterPersonServer(grpcServer, &handler.Person{})
	err := service.Run()
	if err != nil {
		t.Fatal(err)
	}
}

func TestClient(t *testing.T) {
	scheme := fmt.Sprintf("%s_%s", registry.DefaultScheme, registry.DefaultService)
	target := fmt.Sprintf("%s:///", scheme)
	r := registry.NewRegistry(registry.Addrs([]string{"192.168.10.20:2379"}...))
	registry.RegisterResolver(r, registry.ResolverScheme(scheme))

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cancel()

	conn, err := grpc.DialContext(ctx, target, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingPolicy": "%s"}`, roundrobin.Name)))

	if err != nil {
		t.Fatal(err)
	}

	client := person.NewPersonClient(conn)

	ctx, cancel = context.WithTimeout(context.TODO(), time.Second*5)
	defer cancel()

	req := &person.SayHelloRequest{
		Name: "Foo",
	}

	for i := 9; i < 100; i++ {
		rsp, err := client.SayHello(ctx, req)
		if err != nil {
			t.Log("err", err)
		}
		fmt.Println(rsp)
		time.Sleep(time.Second)
	}
}
