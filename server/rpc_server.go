package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/w3liu/bull/codec"
	"github.com/w3liu/bull/infra/addr"
	"github.com/w3liu/bull/infra/backoff"
	mnet "github.com/w3liu/bull/infra/net"
	"github.com/w3liu/bull/infra/socket"
	log "github.com/w3liu/bull/logger"
	"github.com/w3liu/bull/metadata"
	"github.com/w3liu/bull/registry"
	"github.com/w3liu/bull/transport"
	"go.uber.org/zap"
	"net"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	lastStreamResponseError = errors.New("EOS")
)

type rpcServer struct {
	exit chan chan error

	sync.RWMutex
	opts     Options
	handlers map[string]Handler
	// marks the serve as started
	started bool
	// used for first registration
	registered bool
	// graceful exit
	wg *sync.WaitGroup

	rsvc *registry.Service
}

func newRpcServer(opts ...Option) Server {
	options := newOptions(opts...)

	return &rpcServer{
		opts:     options,
		handlers: make(map[string]Handler),
		exit:     make(chan chan error),
		wg:       wait(options.Context),
	}
}

// ServeConn serves a single connection
func (s *rpcServer) ServeConn(sock transport.Socket) {
	// global error tracking
	var gerr error
	// streams are multiplexed on Bull-Stream or Bull-Id header
	pool := socket.NewPool()

	// get global waitgroup
	s.Lock()
	gg := s.wg
	s.Unlock()

	// waitgroup to wait for processing to finish
	wg := &waitGroup{
		gg: gg,
	}

	defer func() {
		// only wait if there's no error
		if gerr == nil {
			// wait till done
			wg.Wait()
		}

		// close all the sockets for this connection
		pool.Close()

		// close underlying socket
		sock.Close()

		// recover any panics
		if r := recover(); r != nil {
			log.Error("panic recovered: ", zap.Any("r", r))
			log.Error(string(debug.Stack()))
		}
	}()

	for {
		var msg transport.Message
		// process inbound messages one at a time
		if err := sock.Recv(&msg); err != nil {
			// set a global error and return
			// we're saying we essentially can't
			// use the socket anymore
			gerr = err
			return
		}

		// business as usual

		// use Bull-Stream as the stream identifier
		// in the event its blank we'll always process
		// on the same socket
		id := msg.Header["Bull-Stream"]

		// if there's no stream id then its a standard request
		// use the Bull-Id
		if len(id) == 0 {
			id = msg.Header["Bull-Id"]
		}

		// check stream id
		var stream bool

		if v := getHeader("Bull-Stream", msg.Header); len(v) > 0 {
			stream = true
		}

		// check if we have an existing socket
		psock, ok := pool.Get(id)

		// if we don't have a socket and its a stream
		if !ok && stream {
			// check if its a last stream EOS error
			err := msg.Header["Bull-Error"]
			if err == lastStreamResponseError.Error() {
				pool.Release(psock)
				continue
			}
		}

		// got an existing socket already
		if ok {
			// we're starting processing
			wg.Add(1)

			// pass the message to that existing socket
			if err := psock.Accept(&msg); err != nil {
				// release the socket if there's an error
				pool.Release(psock)
			}

			// done waiting
			wg.Done()

			// continue to the next message
			continue
		}

		// no socket was found so its new
		// set the local and remote values
		psock.SetLocal(sock.Local())
		psock.SetRemote(sock.Remote())

		// load the socket with the current message
		psock.Accept(&msg)

		// now walk the usual path

		// we use this Timeout header to set a server deadline
		to := msg.Header["Timeout"]
		// we use this Content-Type header to identify the codec needed
		ct := msg.Header["Content-Type"]

		// copy the message headers
		hdr := make(map[string]string, len(msg.Header))
		for k, v := range msg.Header {
			hdr[k] = v
		}

		// set local/remote ips
		hdr["Local"] = sock.Local()
		hdr["Remote"] = sock.Remote()

		// create new context with the metadata
		ctx := metadata.NewContext(context.Background(), hdr)

		// set the timeout from the header if we have it
		if len(to) > 0 {
			if n, err := strconv.ParseUint(to, 10, 64); err == nil {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, time.Duration(n))
				defer cancel()
			}
		}

		// if there's no content type default it
		if len(ct) == 0 {
			msg.Header["Content-Type"] = DefaultContentType
			ct = DefaultContentType
		}

		// setup old protocol
		cf := setupProtocol(&msg)

		// no legacy codec needed
		if cf == nil {
			var err error
			// try get a new codec
			if cf, err = s.newCodec(ct); err != nil {
				// no codec found so send back an error
				if err := sock.Send(&transport.Message{
					Header: map[string]string{
						"Content-Type": "text/plain",
					},
					Body: []byte(err.Error()),
				}); err != nil {
					gerr = err
				}

				// release the socket we just created
				pool.Release(psock)
				// now continue
				continue
			}
		}

		// create a new rpc codec based on the pseudo socket and codec
		rcodec := newRpcCodec(&msg, psock, cf)
		// check the protocol as well
		protocol := rcodec.String()

		// wait for two coroutines to exit
		// serve the request and process the outbound messages
		wg.Add(2)

		// process the outbound messages from the socket
		go func(id string, psock *socket.Socket) {
			defer func() {
				// TODO: don't hack this but if its grpc just break out of the stream
				// We do this because the underlying connection is h2 and its a stream
				switch protocol {
				case "grpc":
					sock.Close()
				}
				// release the socket
				pool.Release(psock)
				// signal we're done
				wg.Done()

				// recover any panics for outbound process
				if r := recover(); r != nil {
					log.Error("panic recovered: ", zap.Any("r", r))
					log.Error(string(debug.Stack()))
				}
			}()

			for {
				// get the message from our internal handler/stream
				m := new(transport.Message)
				if err := psock.Process(m); err != nil {
					return
				}

				// send the message back over the socket
				if err := sock.Send(m); err != nil {
					return
				}
			}
		}(id, psock)

		// serve the request in a go routine as this may be a stream
		go func(id string, psock *socket.Socket) {
			defer func() {
				// release the socket
				pool.Release(psock)
				// signal we're done
				wg.Done()

				// recover any panics for call handler
				if r := recover(); r != nil {
					log.Error("panic recovered: ", zap.Any("r", r))
					log.Error(string(debug.Stack()))
				}
			}()
		}(id, psock)
	}
}

func (s *rpcServer) newCodec(contentType string) (codec.NewCodec, error) {
	if cf, ok := s.opts.Codecs[contentType]; ok {
		return cf, nil
	}
	if cf, ok := DefaultCodecs[contentType]; ok {
		return cf, nil
	}
	return nil, fmt.Errorf("Unsupported Content-Type: %s", contentType)
}

func (s *rpcServer) Options() Options {
	s.RLock()
	opts := s.opts
	s.RUnlock()
	return opts
}

func (s *rpcServer) Init(opts ...Option) error {
	s.Lock()
	defer s.Unlock()

	for _, opt := range opts {
		opt(&s.opts)
	}

	s.rsvc = nil

	return nil
}

func (s *rpcServer) Register() error {
	s.RLock()
	rsvc := s.rsvc
	config := s.Options()
	s.RUnlock()

	regFunc := func(service *registry.Service) error {
		// create registry options
		rOpts := []registry.RegisterOption{registry.RegisterTTL(config.RegisterTTL)}

		var regErr error

		for i := 0; i < 3; i++ {
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

	// have we registered before?
	if rsvc != nil {
		if err := regFunc(rsvc); err != nil {
			return err
		}
		return nil
	}

	var err error
	var advt, host, port string
	var cacheService bool

	// check the advertise address first
	// if it exists then use it, otherwise
	// use the address
	if len(config.Advertise) > 0 {
		advt = config.Advertise
	} else {
		advt = config.Address
	}

	if cnt := strings.Count(advt, ":"); cnt >= 1 {
		// ipv6 address in format [host]:port or ipv4 host:port
		host, port, err = net.SplitHostPort(advt)
		if err != nil {
			return err
		}
	} else {
		host = advt
	}

	if ip := net.ParseIP(host); ip != nil {
		cacheService = true
	}

	addr, err := addr.Extract(host)
	if err != nil {
		return err
	}

	// make copy of metadata
	md := metadata.Copy(config.Metadata)

	// mq-rpc(eg. nats) doesn't need the port. its addr is queue name.
	if port != "" {
		addr = mnet.HostPort(addr, port)
	}

	// register service
	node := &registry.Node{
		Id:       config.Name + "-" + config.Id,
		Address:  addr,
		Metadata: md,
	}

	node.Metadata["transport"] = config.Transport.String()
	node.Metadata["server"] = s.String()
	node.Metadata["registry"] = config.Registry.String()
	node.Metadata["protocol"] = "mucp"

	s.RLock()

	// Maps are ordered randomly, sort the keys for consistency
	var handlerList []string
	for n, e := range s.handlers {
		// Only advertise non internal handlers
		if !e.Options().Internal {
			handlerList = append(handlerList, n)
		}
	}

	sort.Strings(handlerList)

	endpoints := make([]*registry.Endpoint, 0, len(handlerList))

	for _, n := range handlerList {
		endpoints = append(endpoints, s.handlers[n].Endpoints()...)
	}

	service := &registry.Service{
		Name:      config.Name,
		Version:   config.Version,
		Nodes:     []*registry.Node{node},
		Endpoints: endpoints,
	}

	// get registered value
	registered := s.registered

	s.RUnlock()

	if !registered {
		log.Infof("Registry [%s] Registering node: %s", config.Registry.String(), node.Id)
	}

	// register the service
	if err := regFunc(service); err != nil {
		return err
	}

	// already registered? don't need to register subscribers
	if registered {
		return nil
	}

	s.Lock()
	defer s.Unlock()

	// set what we're advertising
	s.opts.Advertise = addr

	if cacheService {
		s.rsvc = service
	}
	s.registered = true

	return nil
}

func (s *rpcServer) Deregister() error {
	var err error
	var advt, host, port string

	s.RLock()
	config := s.Options()
	s.RUnlock()

	// check the advertise address first
	// if it exists then use it, otherwise
	// use the address
	if len(config.Advertise) > 0 {
		advt = config.Advertise
	} else {
		advt = config.Address
	}

	if cnt := strings.Count(advt, ":"); cnt >= 1 {
		// ipv6 address in format [host]:port or ipv4 host:port
		host, port, err = net.SplitHostPort(advt)
		if err != nil {
			return err
		}
	} else {
		host = advt
	}

	addr, err := addr.Extract(host)
	if err != nil {
		return err
	}

	// mq-rpc(eg. nats) doesn't need the port. its addr is queue name.
	if port != "" {
		addr = mnet.HostPort(addr, port)
	}

	node := &registry.Node{
		Id:      config.Name + "-" + config.Id,
		Address: addr,
	}

	service := &registry.Service{
		Name:    config.Name,
		Version: config.Version,
		Nodes:   []*registry.Node{node},
	}

	log.Infof("Registry [%s] Deregistering node: %s", config.Registry.String(), node.Id)

	if err := config.Registry.Deregister(service); err != nil {
		return err
	}

	s.Lock()
	s.rsvc = nil

	if !s.registered {
		s.Unlock()
		return nil
	}

	s.registered = false

	s.Unlock()
	return nil
}

func (s *rpcServer) Start() error {
	s.RLock()
	if s.started {
		s.RUnlock()
		return nil
	}
	s.RUnlock()

	config := s.Options()

	// start listening on the transport
	ts, err := config.Transport.Listen(config.Address)
	if err != nil {
		return err
	}

	log.Infof("Transport [%s] Listening on %s", config.Transport.String(), ts.Addr())

	// swap address
	s.Lock()
	addr := s.opts.Address
	s.opts.Address = ts.Addr()
	s.Unlock()

	// use RegisterCheck func before register
	if err = s.opts.RegisterCheck(s.opts.Context); err != nil {

		log.Errorf("Server %s-%s register check error: %s", config.Name, config.Id, err)
	} else {
		// announce self to the world
		if err = s.Register(); err != nil {
			log.Errorf("Server %s-%s register error: %s", config.Name, config.Id, err)
		}
	}

	exit := make(chan bool)

	go func() {
		for {
			// listen for connections
			err := ts.Accept(s.ServeConn)

			// TODO: listen for messages
			// msg := broker.Exchange(service).Consume()

			select {
			// check if we're supposed to exit
			case <-exit:
				return
			// check the error and backoff
			default:
				if err != nil {
					log.Errorf("Accept error: %v", err)
					time.Sleep(time.Second)
					continue
				}
			}

			// no error just exit
			return
		}
	}()

	go func() {
		t := new(time.Ticker)

		// only process if it exists
		if s.opts.RegisterInterval > time.Duration(0) {
			// new ticker
			t = time.NewTicker(s.opts.RegisterInterval)
		}

		// return error chan
		var ch chan error

	Loop:
		for {
			select {
			// register self on interval
			case <-t.C:
				s.RLock()
				registered := s.registered
				s.RUnlock()
				rerr := s.opts.RegisterCheck(s.opts.Context)
				if rerr != nil && registered {
					log.Errorf("Server %s-%s register check error: %s, deregister it", config.Name, config.Id, err)
					// deregister self in case of error
					if err := s.Deregister(); err != nil {
						log.Errorf("Server %s-%s deregister error: %s", config.Name, config.Id, err)
					}
				} else if rerr != nil && !registered {
					log.Errorf("Server %s-%s register check error: %s", config.Name, config.Id, err)
					continue
				}
				if err := s.Register(); err != nil {
					log.Errorf("Server %s-%s register error: %s", config.Name, config.Id, err)
				}
			// wait for exit
			case ch = <-s.exit:
				t.Stop()
				close(exit)
				break Loop
			}
		}

		s.RLock()
		registered := s.registered
		s.RUnlock()
		if registered {
			// deregister self
			if err := s.Deregister(); err != nil {
				log.Errorf("Server %s-%s deregister error: %s", config.Name, config.Id, err)
			}
		}

		s.Lock()
		swg := s.wg
		s.Unlock()

		// wait for requests to finish
		if swg != nil {
			swg.Wait()
		}

		// close transport listener
		ch <- ts.Close()

		// swap back address
		s.Lock()
		s.opts.Address = addr
		s.Unlock()
	}()

	// mark the server as started
	s.Lock()
	s.started = true
	s.Unlock()

	return nil
}

func (s *rpcServer) Stop() error {
	s.RLock()
	if !s.started {
		s.RUnlock()
		return nil
	}
	s.RUnlock()

	ch := make(chan error)
	s.exit <- ch

	err := <-ch
	s.Lock()
	s.started = false
	s.Unlock()

	return err
}

func (s *rpcServer) String() string {
	return "mucp"
}
