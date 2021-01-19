package server

import (
	"github.com/w3liu/bull/infra/addr"
	"github.com/w3liu/bull/infra/backoff"
	mnet "github.com/w3liu/bull/infra/net"
	"github.com/w3liu/bull/logger"
	meta "github.com/w3liu/bull/metadata"
	"github.com/w3liu/bull/registry"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
	"strings"
	"sync"
	"time"
)

type grpcServer struct {
	srv  *grpc.Server
	exit chan chan error
	wg   *sync.WaitGroup

	sync.RWMutex
	opts       Options
	started    bool
	registered bool
	rsvc       *registry.Service
}

func newServer(opts ...Option) Server {
	options := newOptions(opts...)

	srv := &grpcServer{
		srv:        nil,
		exit:       make(chan chan error),
		wg:         wait(options.Context),
		RWMutex:    sync.RWMutex{},
		opts:       options,
		started:    false,
		registered: false,
		rsvc:       nil,
	}

	srv.configure()

	return srv
}

func (g *grpcServer) Init(opts ...Option) error {
	g.configure(opts...)
	return nil
}

func (g *grpcServer) configure(opts ...Option) {
	g.Lock()
	defer g.Unlock()

	// Don't reprocess where there's no config
	if len(opts) == 0 && g.srv != nil {
		return
	}

	for _, o := range opts {
		o(&g.opts)
	}

	maxMsgSize := DefaultMaxMsgSize

	gopts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(maxMsgSize),
		grpc.MaxSendMsgSize(maxMsgSize),
	}

	g.rsvc = nil
	g.srv = grpc.NewServer(gopts...)
}

func (g *grpcServer) Options() Options {

	g.RLock()
	opts := g.opts
	g.RUnlock()

	return opts
}

func (g *grpcServer) Start() error {
	g.RLock()
	if g.started {
		g.RUnlock()
		return nil
	}
	g.RUnlock()
	config := g.Options()

	ts, err := net.Listen("tcp", config.Address)
	if err != nil {
		return err
	}

	logger.Infof("Server [grpc] Listening on %s", ts.Addr().String())

	g.Lock()
	g.opts.Address = ts.Addr().String()
	g.Unlock()

	if err := g.Register(); err != nil {
		return err
	}

	go func() {
		if err := g.srv.Serve(ts); err != nil {
			logger.Errorf("gRPC Server start error: %v", err)
		}
	}()

	go func() {
		t := new(time.Ticker)

		// only process if it exists
		if g.opts.RegisterInterval > time.Duration(0) {
			// new ticker
			t = time.NewTicker(g.opts.RegisterInterval)
		}

		// return error chan
		var ch chan error

	Loop:
		for {
			select {
			// register self on interval
			case <-t.C:
				if err := g.Register(); err != nil {
					logger.Error("Server register error: ", zap.Error(err))
				}
			// wait for exit
			case ch = <-g.exit:
				break Loop
			}
		}

		// deregister self
		if err := g.Deregister(); err != nil {
			logger.Error("Server deregister error: ", zap.Error(err))
		}

		// wait for waitgroup
		if g.wg != nil {
			g.wg.Wait()
		}

		// stop the grpc server
		exit := make(chan bool)

		go func() {
			g.srv.GracefulStop()
			close(exit)
		}()

		select {
		case <-exit:
		case <-time.After(time.Second):
			g.srv.Stop()
		}

		// close transport
		ch <- nil
	}()

	g.Lock()
	g.started = true
	g.Unlock()

	return nil
}

func (g *grpcServer) Stop() error {
	g.RLock()
	if !g.started {
		g.RUnlock()
		return nil
	}
	g.RUnlock()

	ch := make(chan error)
	g.exit <- ch

	var err error
	select {
	case err = <-ch:
		g.Lock()
		g.rsvc = nil
		g.started = false
		g.Unlock()
	}

	return err
}

func (g *grpcServer) String() string {
	return "grpc"
}

func (g *grpcServer) Register() error {
	g.RLock()
	rsvc := g.rsvc
	config := g.opts
	g.RUnlock()

	regFunc := func(service *registry.Service) error {
		var regErr error

		for i := 0; i < 3; i++ {
			// set the ttl
			rOpts := []registry.RegisterOption{registry.RegisterTTL(config.RegisterTTL)}
			// attempt to register
			if err := config.Registry.Register(service, rOpts...); err != nil {
				// set the error
				regErr = err
				// backoff then retry
				time.Sleep(backoff.Do(i + 1))
				continue
			}
			// success so nil error
			regErr = nil
			break
		}

		return regErr
	}

	// if service already filled, reuse it and return early
	if rsvc != nil {
		if err := regFunc(rsvc); err != nil {
			return err
		}
		return nil
	}

	var err error
	var address, host, port string
	var cacheService bool

	address = config.Address

	if cnt := strings.Count(address, ":"); cnt >= 1 {
		// ipv6 address in format [host]:port or ipv4 host:port
		host, port, err = net.SplitHostPort(address)
		if err != nil {
			return err
		}
	} else {
		host = address
	}

	if ip := net.ParseIP(host); ip != nil {
		cacheService = true
	}

	exAddr, err := addr.Extract(host)
	if err != nil {
		return err
	}

	md := meta.Copy(config.Metadata)

	node := &registry.Node{
		Id:       config.Name + "-" + config.Id,
		Address:  mnet.HostPort(exAddr, port),
		Metadata: md,
	}

	node.Metadata["registry"] = config.Registry.String()
	node.Metadata["server"] = g.String()
	node.Metadata["transport"] = g.String()
	node.Metadata["protocol"] = "grpc"

	service := &registry.Service{
		Name:    config.Name,
		Version: config.Version,
		Nodes:   []*registry.Node{node},
	}

	g.RLock()
	registered := g.registered
	g.RUnlock()

	if !registered {
		logger.Infof("Registry [%s] Registering node: %s", config.Registry.String(), node.Id)
	}

	// register the service
	if err := regFunc(service); err != nil {
		return err
	}

	// already registered? don't need to register subscribers
	if registered {
		return nil
	}

	g.Lock()
	defer g.Unlock()

	g.registered = true
	if cacheService {
		g.rsvc = service
	}

	return nil
}

func (g *grpcServer) Deregister() error {
	var err error
	var address, host, port string

	g.RLock()
	config := g.opts
	g.RUnlock()

	// check the advertise address first
	// if it exists then use it, otherwise
	// use the address
	address = config.Address

	if cnt := strings.Count(address, ":"); cnt >= 1 {
		// ipv6 address in format [host]:port or ipv4 host:port
		host, port, err = net.SplitHostPort(address)
		if err != nil {
			return err
		}
	} else {
		host = address
	}

	exAddr, err := addr.Extract(host)
	if err != nil {
		return err
	}

	node := &registry.Node{
		Id:      config.Name + "-" + config.Id,
		Address: mnet.HostPort(exAddr, port),
	}

	service := &registry.Service{
		Name:    config.Name,
		Version: config.Version,
		Nodes:   []*registry.Node{node},
	}
	logger.Infof("Deregistering node: %s", node.Id)
	if err := config.Registry.Deregister(service); err != nil {
		return err
	}

	g.Lock()
	g.rsvc = nil

	if !g.registered {
		g.Unlock()
		return nil
	}

	g.registered = false

	g.Unlock()
	return nil
}

func (g *grpcServer) Instance() interface{} {
	return g.srv
}
