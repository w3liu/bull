package bull

import (
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
	person.RegisterPersonServer(server.Instance().(*grpc.Server), &handler.Person{})
	err := service.Run()
	if err != nil {
		t.Fatal(err)
	}
}
