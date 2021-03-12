package bull

import (
	"context"
	"fmt"
	"github.com/w3liu/bull/client"
	"github.com/w3liu/bull/examples/handler"
	pb "github.com/w3liu/bull/examples/proto"
	"github.com/w3liu/bull/registry"
	"github.com/w3liu/bull/server"
	"google.golang.org/grpc"
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

func TestService3(t *testing.T) {
	err := runService("Foo3")
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
	grpcServer, ok := serv.Instance().(*grpc.Server)
	if !ok {
		panic("not grpc server")
	}
	pb.RegisterPersonServer(grpcServer, &handler.Person{Name: name})
	err := service.Run()
	return err
}

func TestClient(t *testing.T) {
	r := registry.NewRegistry(registry.Addrs([]string{"127.0.0.1:2379"}...))
	cli := client.NewClient(
		client.Registry(r),
		client.Service("hello.svc"))

	conn, ok := cli.Instance().(*grpc.ClientConn)
	if !ok {
		panic("not grpc client conn instance")
	}
	personClient := pb.NewPersonClient(conn)

	req := &pb.SayHelloRequest{
		Name: "Bar",
	}

	for i := 0; i < 100; i++ {
		ctx, _ := context.WithTimeout(context.TODO(), time.Second*5)
		rsp, err := personClient.SayHello(ctx, req)
		if err != nil {
			t.Log("err", err)
		}
		fmt.Println(rsp)
		time.Sleep(time.Second)
	}
}
