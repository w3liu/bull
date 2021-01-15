package registry

import (
	"github.com/w3liu/bull/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc/resolver"
	"time"
)

var (
	DefaultScheme  = "etcd"
	DefaultService = "go.bull.server"
	DefaultTimeOut = time.Second * 5
)

type etcdResolver struct {
	registry *etcdRegistry
	scheme   string
	service  string
	timeOut  time.Duration
	watcher  Watcher
	target   resolver.Target
	cc       resolver.ClientConn
}

func RegisterResolver(r *etcdRegistry, opts ...ResolverOption) {
	options := newResolverOptions()

	for _, o := range opts {
		o(&options)
	}

	resolver.Register(&etcdResolver{
		registry: r,
		scheme:   options.Scheme,
		service:  options.Service,
		timeOut:  options.TimeOut,
	})
}

func newResolverOptions() ResolverOptions {
	return ResolverOptions{
		Scheme:  DefaultScheme,
		Service: DefaultService,
		TimeOut: DefaultTimeOut,
	}
}

func (r *etcdResolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r.target = target
	r.cc = cc
	watcher, err := newEtcdWatcher(r.registry, r.timeOut, WatchService(r.service))
	if err != nil {
		return nil, err
	}
	r.watcher = watcher
	r.start()
	return r, nil
}

func (r *etcdResolver) Scheme() string {
	return r.scheme
}

func (r *etcdResolver) ResolveNow(o resolver.ResolveNowOptions) {
}

func (r *etcdResolver) Close() {
	r.watcher.Stop()
}

func (r *etcdResolver) start() {
	go func() {
		for {
			result, err := r.watcher.Next()
			if err != nil {
				logger.Error("watcher next error", zap.Error(err))
				return
			}
			var addrs = make([]resolver.Address, 0)
			if result != nil && result.Service != nil {
				nodes := result.Service.Nodes
				for _, node := range nodes {
					addrs = append(addrs, resolver.Address{Addr: node.Address})
				}
				r.cc.UpdateState(resolver.State{Addresses: addrs})
			}
		}
	}()
}
