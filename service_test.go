package bull

import (
	"context"
	"github.com/w3liu/bull/debug/handler"
	"github.com/w3liu/bull/debug/proto/person"
	"github.com/w3liu/bull/registry"
	"google.golang.org/grpc"
	"testing"
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
	conn, err := grpc.Dial(":53505", grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}

	client := person.NewPersonClient(conn)
	rsp, err := client.SayHello(context.TODO(), &person.SayHelloRequest{
		Name: "Foo",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(rsp)
}
