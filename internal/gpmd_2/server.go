package gpmd

import (
	"crypto/tls"
	"log/slog"
	"net"
	"net/rpc"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/debug"
	"github.com/Morgahl/gotp/internal/ctx"

	icrypto "github.com/Morgahl/gotp/internal/crypto"
)

type Server struct {
	gpmd      *GPMD
	rpcServer *rpc.Server
	cert      tls.Certificate
	listener  net.Listener
}

func NewServer(ctx ctx.Cancellable) *Server {
	rpcServer := rpc.NewServer()
	gpmd := New(ctx)
	debug.AssertNil(rpcServer.Register(gpmd), "Failed to register GPMD")

	nonce, err := icrypto.GenerateNonce()
	debug.AssertNil(err, "Failed to generate nonce: %s", err)

	cert, err := icrypto.GenerateSelfSignedCert("gpmd", gotp.Atom(string(nonce[:])))
	debug.AssertNil(err, "Failed to generate self-signed certificate: %s", err)

	s := &Server{
		gpmd:      gpmd,
		rpcServer: rpcServer,
		cert:      cert,
	}

	return s
}

func (s *Server) Start() {
	debug.AssertNil(s.listener, "Server already started")
	slog.Info("Starting GPMD server")
	debug.AssertNil(s.listener, "Server already started")
	var err error
	s.listener, err = icrypto.Listen(DEFAULT_BIND, s.cert)
	debug.AssertNil(err, "Failed to start listener: %s", err)
	go s.listen()
}

func (s *Server) Stop() {
	debug.AssertNotNil(s.listener, "Server not started")
	debug.AssertNil(s.listener.Close(), "Failed to close listener")
	slog.Info("Stopping GPMD server")
	s.listener = nil
}

func (s *Server) listen() {
	slog.Info("GPMD server listening", "address", s.listener.Addr())
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			slog.Debug("closing listener", "error", err)
			return
		}
		// TODO: FUTURE HOME OF A LINKED PROCESS SPAWN INSTEAD OF THIS GO ROUTINE
		go s.rpcServer.ServeConn(conn)
	}
}
