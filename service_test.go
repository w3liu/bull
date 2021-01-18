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
	r := registry.NewRegistry(registry.Addrs([]string{"192.168.10.20:2379"}...))
	registry.RegisterResolver(r)

	conn, err := grpc.Dial(fmt.Sprintf("%s:///", registry.DefaultScheme), grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, roundrobin.Name)))

	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(30 * time.Second)

	client := person.NewPersonClient(conn)

	ctx, _ := context.WithTimeout(context.TODO(), time.Second*5)

	rsp, err := client.SayHello(ctx, &person.SayHelloRequest{
		Name: "Foo",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(rsp)
}
