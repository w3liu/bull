package bull

import (
	"context"
	"errors"
	"github.com/w3liu/bull/client"
	"github.com/w3liu/bull/client/grpc"
	"github.com/w3liu/bull/debug/handler"
	proto "github.com/w3liu/bull/debug/proto"
	"github.com/w3liu/bull/debug/proto/person"
	"github.com/w3liu/bull/registry"
	"github.com/w3liu/bull/server"
	"sync"
	"testing"
	"time"
)

func testShutdown(wg *sync.WaitGroup, cancel func()) {
	// add 1
	wg.Add(1)
	// shutdown the service
	cancel()
	// wait for stop
	wg.Wait()
}

func testService(ctx context.Context, wg *sync.WaitGroup, name string) Service {
	// add self
	wg.Add(1)

	r := registry.NewRegistry(registry.Addrs([]string{"192.168.10.20:2379"}...))

	// create service
	return NewService(
		Name(name),
		Context(ctx),
		Server(server.NewServer(server.Registry(r), server.Name(name))),
		Client(grpc.NewClient(client.Registry(r))),
		AfterStart(func() error {
			wg.Done()
			return nil
		}),
		AfterStop(func() error {
			wg.Done()
			return nil
		}),
	)
}

func TestService(t *testing.T) {
	// waitgroup for server start
	var wg sync.WaitGroup

	// cancellation context
	ctx, cancel := context.WithCancel(context.Background())

	// start test server
	service := testService(ctx, &wg, "test.service")

	go func() {
		// wait for service to start
		wg.Wait()

		// make a test call
		if err := testRequest(ctx, service.Client(), "test.service"); err != nil {
			t.Fatal(err)
		}

		// shutdown the service
		testShutdown(&wg, cancel)
	}()

	// register the debug handler
	service.Server().Handle(
		service.Server().NewHandler(
			handler.NewHandler(),
			server.InternalHandler(false),
		),
	)

	// start service
	if err := service.Run(); err != nil {
		t.Fatal(err)
	}
}

func testRequest(ctx context.Context, c client.Client, name string) error {
	// test call debug
	req := c.NewRequest(
		name,
		"Debug.Health",
		new(proto.HealthRequest),
	)

	rsp := new(proto.HealthResponse)

	time.Sleep(time.Second * 10)

	err := c.Call(context.TODO(), req, rsp)
	if err != nil {
		return err
	}

	if rsp.Status != "ok" {
		return errors.New("service response: " + rsp.Status)
	}

	return nil
}

func TestService1(t *testing.T) {
	var wg sync.WaitGroup

	// cancellation context
	ctx, cancel := context.WithCancel(context.Background())

	// start test server
	service := testService(ctx, &wg, "test.service")

	go func() {
		// wait for service to start
		wg.Wait()

		// make a test call
		if err := testRequest1(ctx, service.Client(), "test.service"); err != nil {
			t.Fatal(err)
		}

		// shutdown the service
		testShutdown(&wg, cancel)
	}()

	// register the debug handler
	service.Server().Handle(
		service.Server().NewHandler(
			handler.NewPersonService("w3liu"),
			server.InternalHandler(false),
		),
	)

	// start service
	if err := service.Run(); err != nil {
		t.Fatal(err)
	}
}

func testRequest1(ctx context.Context, c client.Client, name string) error {
	// test call debug
	req := c.NewRequest(
		name,
		"Person.SayHello",
		&person.SayHelloRequest{
			Name: "Thomas",
		},
	)

	rsp := new(person.SayHelloResponse)

	time.Sleep(time.Second * 10)

	err := c.Call(context.TODO(), req, rsp)
	if err != nil {
		return err
	}

	if len(rsp.Msg) == 0 {
		return errors.New("len(rsp.Msg) == 0")
	}

	return nil
}
