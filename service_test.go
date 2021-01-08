package bull

import (
	"context"
	"errors"
	"github.com/w3liu/bull/client"
	proto "github.com/w3liu/bull/debug/proto"
	"github.com/w3liu/bull/registry"
	"sync"
	"testing"
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

	r := registry.NewRegistry()

	// create service
	return NewService(
		Name(name),
		Context(ctx),
		Registry(r),
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

	err := c.Call(context.TODO(), req, rsp)
	if err != nil {
		return err
	}

	if rsp.Status != "ok" {
		return errors.New("service response: " + rsp.Status)
	}

	return nil
}
