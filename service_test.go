package bull

import (
	"context"
	"fmt"
	"github.com/w3liu/bull/debug/handler"
	"github.com/w3liu/bull/debug/proto/person"
	"github.com/w3liu/bull/registry"
	"github.com/w3liu/bull/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"testing"
	"time"
)

func TestService1(t *testing.T) {
	err := runService("Foo1")
	if err != nil {
		t.Fatal(err)
	}
}

func TestService2(t *testing.T) {
	err := runService("Foo2")
	if err != nil {
		t.Fatal(err)
	}
}

func runService(name string) error {
	r := registry.NewRegistry(registry.Addrs([]string{"127.0.0.1:2379"}...))
	service := NewService(
		Registry(r),
		Server(server.NewServer(server.Name(fmt.Sprintf("%s_%d", "hello.svc", 0)))),
	)
	serv := service.Server()
	grpcServer := serv.Instance().(*grpc.Server)
	person.RegisterPersonServer(grpcServer, &handler.Person{Name: name})
	err := service.Run()
	return err
}

func TestClient(t *testing.T) {
	scheme := fmt.Sprintf("%s", registry.DefaultScheme)
	target := fmt.Sprintf("%s:///", scheme)
	r := registry.NewRegistry(registry.Addrs([]string{"127.0.0.1:2379"}...))
	registry.RegisterResolver(r, registry.ResolverScheme(scheme), registry.ResolverService(fmt.Sprintf("%s_%d", "hello.svc", 0)))

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)
	defer cancel()

	conn, err := grpc.DialContext(ctx, target, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingPolicy": "%s"}`, roundrobin.Name)))

	if err != nil {
		t.Fatal(err)
	}

	client := person.NewPersonClient(conn)

	req := &person.SayHelloRequest{
		Name: "Bar",
	}

	for i := 0; i < 100; i++ {
		ctx, _ = context.WithTimeout(context.TODO(), time.Second*5)
		rsp, err := client.SayHello(ctx, req)
		if err != nil {
			t.Log("err", err)
		}
		fmt.Println(rsp)
		time.Sleep(time.Second)
	}
}
