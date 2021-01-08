package transport

import (
	"github.com/w3liu/bull/errors"
	pb "github.com/w3liu/bull/transport/proto"
	"google.golang.org/grpc/peer"
)

type bullTransport struct {
	addr string
	fn   func(Socket)
}

func (m *bullTransport) Stream(ts pb.Transport_StreamServer) (err error) {

	sock := &grpcTransportSocket{
		stream: ts,
		local:  m.addr,
	}

	p, ok := peer.FromContext(ts.Context())
	if ok {
		sock.remote = p.Addr.String()
	}

	defer func() {
		if r := recover(); r != nil {
			//logger.Error(r, string(debug.Stack()))
			sock.Close()
			err = errors.InternalServerError("go.bull.transport", "panic recovered: %v", r)
		}
	}()

	// execute socket func
	m.fn(sock)

	return err
}
