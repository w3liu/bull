package server

import (
	"context"
	"crypto/tls"
	"github.com/w3liu/bull/codec"
	"github.com/w3liu/bull/registry"
	"github.com/w3liu/bull/transport"
	"google.golang.org/grpc/encoding"
	"net"
)

type codecsKey struct{}
type grpcOptions struct{}
type netListener struct{}
type maxMsgSizeKey struct{}
type maxConnKey struct{}
type tlsAuth struct{}

// gRPC Codec to be used to encode/decode requests for a given content type
func GRPCCodec(contentType string, c encoding.Codec) Option {
	return func(o *Options) {
		codecs := make(map[string]encoding.Codec)
		if o.Context == nil {
			o.Context = context.Background()
		}
		if v, ok := o.Context.Value(codecsKey{}).(map[string]encoding.Codec); ok && v != nil {
			codecs = v
		}
		codecs[contentType] = c
		o.Context = context.WithValue(o.Context, codecsKey{}, codecs)
	}
}

// AuthTLS should be used to setup a secure authentication using TLS
func AuthTLS(t *tls.Config) Option {
	return setServerOption(tlsAuth{}, t)
}

// MaxConn specifies maximum number of max simultaneous connections to server
func MaxConn(n int) Option {
	return setServerOption(maxConnKey{}, n)
}

// Listener specifies the net.Listener to use instead of the default
func Listener(l net.Listener) Option {
	return setServerOption(netListener{}, l)
}

// Options to be used to configure gRPC options
//func Options(opts ...grpc.ServerOption) server.Option {
//	return setServerOption(grpcOptions{}, opts)
//}

//
// MaxMsgSize set the maximum message in bytes the server can receive and
// send.  Default maximum message size is 4 MB.
//
func MaxMsgSize(s int) Option {
	return setServerOption(maxMsgSizeKey{}, s)
}

func newGRPCOptions(opt ...Option) Options {
	opts := Options{
		Codecs:    make(map[string]codec.NewCodec),
		Metadata:  map[string]string{},
		Registry:  registry.DefaultRegistry,
		Transport: transport.DefaultTransport,
		Address:   DefaultAddress,
		Name:      DefaultName,
		Id:        DefaultId,
		Version:   DefaultVersion,
	}

	for _, o := range opt {
		o(&opts)
	}

	return opts
}
