package web

import (
	"context"
	"github.com/w3liu/bull/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type service struct {
	opts   Options
	server *http.Server
}

func newService(opts ...Option) Service {
	options := newOptions(opts...)
	return &service{opts: options}
}

func (s *service) Name() string {
	return s.opts.Name
}

func (s *service) Init(opts ...Option) {
	for _, o := range opts {
		o(&s.opts)
	}
	server := &http.Server{Addr: s.opts.Address}
	s.server = server
}

func (s *service) Options() Options {
	return s.opts
}

func (s *service) Handle(pattern string, handler http.Handler) {
	http.Handle(pattern, handler)
}

func (s *service) Run() error {

	logger.Infof("Starting [service] %s", s.Name())
	logger.Infof("Listen at %s", s.opts.Address)

	if err := s.Start(); err != nil {
		return err
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL)

	select {
	// wait on kill signal
	case <-ch:
	// wait on context cancel
	case <-s.opts.Context.Done():
	}

	return s.Stop()

}

func (s *service) String() string {
	return "bull"
}

func (s *service) Start() error {
	go func() {
		if err := s.server.ListenAndServe(); err != nil {
			logger.Errorf("bull web start error: %v", err)
		}
	}()
	return nil
}

func (s *service) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err := s.server.Shutdown(ctx)
	return err
}
