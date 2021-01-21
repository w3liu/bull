package registry

import (
	"github.com/w3liu/bull/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc/resolver"
	"sync"
	"time"
)

var (
	DefaultScheme  = "etcd"
	DefaultService = "go.bull.server"
	DefaultTimeOut = time.Second * 5
)

type etcdResolver struct {
	sync.Mutex
	registry Registry
	scheme   string
	service  string
	timeOut  time.Duration
	watcher  Watcher
	target   resolver.Target
	cc       resolver.ClientConn
	svcs     map[string]*Service
}

func RegisterResolver(r Registry, opts ...ResolverOption) {
	options := newResolverOptions()

	for _, o := range opts {
		o(&options)
	}

	resolver.Register(&etcdResolver{
		registry: r,
		scheme:   options.Scheme,
		timeOut:  options.TimeOut,
		svcs:     make(map[string]*Service),
	})
}

func newResolverOptions() ResolverOptions {
	return ResolverOptions{
		Scheme:  DefaultScheme,
		TimeOut: DefaultTimeOut,
	}
}

func (r *etcdResolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r.target = target
	r.cc = cc
	er := r.registry.(*etcdRegistry)
	watcher, err := newEtcdWatcher(er, r.timeOut)
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
	services, err := r.registry.ListServices()
	if err != nil {
		logger.Error("r.registry.ListServices() error", zap.Error(err))
		return
	}
	for _, service := range services {
		if service != nil {
			r.updateState(service.Name, service.Nodes)
			r.Lock()
			r.svcs[service.Name] = service
			r.Unlock()
		}
	}
	go func() {
		for {
			result, err := r.watcher.Next()
			if err != nil {
				logger.Error("watcher next error", zap.Error(err))
				return
			}
			if result != nil && result.Service != nil {
				res := result.Service
				var nodes []*Node
				switch result.Action {
				case "create", "set":
					nodes = r.addNode(res)
				case "update":
					nodes = r.updateNode(res)
				case "delete", "expire":
					nodes = r.deleteNode(res)
				default:
					logger.Warn("watcher next unknown action", zap.Any("action", result.Action))
					continue
				}
				r.Lock()
				svc, ok := r.svcs[res.Name]
				if !ok {
					svc = res
				}
				if len(nodes) == 0 {
					delete(r.svcs, res.Name)
				} else {
					r.svcs[res.Name] = &Service{
						Name:     svc.Name,
						Version:  svc.Version,
						Metadata: svc.Metadata,
						Nodes:    nodes,
					}
				}
				r.Unlock()
				r.updateState(svc.Name, nodes)
			}
		}
	}()
}

func (r *etcdResolver) updateState(name string, nodes []*Node) {
	var addrs = make([]resolver.Address, 0)
	for _, node := range nodes {
		addrs = append(addrs, resolver.Address{Addr: node.Address, ServerName: name})
	}
	r.cc.UpdateState(resolver.State{Addresses: addrs})
}

func (r *etcdResolver) addNode(res *Service) []*Node {
	var nodes []*Node
	r.Lock()
	svc, ok := r.svcs[res.Name]
	r.Unlock()
	if ok {
		nodes = svc.Nodes
	} else {
		nodes = make([]*Node, 0)
	}
	for _, add := range res.Nodes {
		var exist bool
		for _, cur := range nodes {
			if cur.Id == add.Id {
				exist = true
				break
			}
		}
		if !exist {
			nodes = append(nodes, add)
		}
	}
	return nodes
}

func (r *etcdResolver) updateNode(res *Service) []*Node {
	var nodes []*Node
	r.Lock()
	svc, ok := r.svcs[res.Name]
	r.Unlock()
	if ok {
		nodes = svc.Nodes
	} else {
		nodes = make([]*Node, 0)
	}
	for _, update := range res.Nodes {
		for i, cur := range nodes {
			if cur.Id == update.Id {
				nodes[i] = update
			}
		}
	}
	return nodes
}

func (r *etcdResolver) deleteNode(res *Service) []*Node {
	var nodes []*Node
	r.Lock()
	svc, ok := r.svcs[res.Name]
	r.Unlock()
	if ok {
		nodes = svc.Nodes
	} else {
		nodes = make([]*Node, 0)
	}
	for i, cur := range nodes {
		var exist bool
		for _, del := range res.Nodes {
			if cur.Id == del.Id {
				exist = true
				break
			}
		}
		if exist {
			nodes = append(nodes[:i], nodes[i+1:]...)
		}
	}
	return nodes
}
