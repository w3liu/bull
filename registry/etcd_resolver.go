package registry

import (
	"google.golang.org/grpc/resolver"
	"sync"
	"time"
)

type etcdResolver struct {
	scheme  string
	watcher Watcher
	target  resolver.Target
	cc      resolver.ClientConn
	wg      sync.WaitGroup
}

func NewResolver(r *etcdRegistry) (*etcdResolver, error) {
	watcher, err := newEtcdWatcher(r, time.Second*10)
	if err != nil {
		return nil, err
	}
	return &etcdResolver{
		watcher: watcher,
	}, err
}

func (r *etcdResolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r.target = target
	r.cc = cc
	return r, nil
}

func (r *etcdResolver) Scheme() string {
	return r.scheme
}

func (r *etcdResolver) ResolveNow(o resolver.ResolveNowOptions) {
}

func (r *etcdResolver) Close() {

	r.wg.Wait()
}

func (r *etcdResolver) start() {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		//result, err := r.watcher.Next()
		//if err != nil {
		//
		//}
		//
		//r.cc.UpdateState(resolver.State{Addresses: addr})
	}()
}
