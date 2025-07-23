package gpmd

import (
	"crypto/tls"
	"log/slog"
	"net"
	"net/rpc"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/assert"
	"github.com/Morgahl/gotp/internal/ctx"
	"github.com/Morgahl/gotp/process"

	icrypto "github.com/Morgahl/gotp/internal/crypto"
)

type Server struct {
	ctx       ctx.Cancellable
	gpmd      *GPMD
	rpcServer *rpc.Server
	cert      tls.Certificate
	listener  net.Listener
}

func NewServer(ctx ctx.Cancellable) *Server {
	rpcServer := rpc.NewServer()
	gpmd := New(ctx)
	assert.Nil(rpcServer.Register(gpmd), "Failed to register GPMD")
	nonce := assert.OkF(icrypto.GenerateNonce())("Failed to generate nonce: %s")
	cert := assert.OkF(icrypto.GenerateSelfSignedCert("gpmd", gotp.Atom(string(nonce[:]))))("Failed to generate self-signed certificate: %s")

	s := &Server{
		ctx:       ctx,
		gpmd:      gpmd,
		rpcServer: rpcServer,
		cert:      cert,
	}

	return s
}

func (s *Server) Start() {
	slog.Info("Starting GPMD server")
	assert.Nil(s.listener, "Server already started")
	s.listener = assert.OkF(icrypto.Listen(DEFAULT_BIND, s.cert))("Failed to start listener: %s")
	s.listen()
}

func (s *Server) Stop() {
	assert.NotNil(s.listener, "Server not started")
	assert.Nil(s.listener.Close(), "Failed to close listener")
	slog.Info("Stopping GPMD server")
	s.ctx.Cancel(process.NORMAL)
}

func (s *Server) listen() {
	var err error
	defer s.ctx.Cancel(err)
	slog.Info("GPMD server listening", "address", s.listener.Addr())
	for {
		var conn net.Conn
		conn, err = s.listener.Accept()
		if err != nil {
			slog.Debug("closing listener", "error", err)
			return
		}
		// TODO: FUTURE HOME OF A LINKED PROCESS SPAWN INSTEAD OF THIS GO ROUTINE
		go s.rpcServer.ServeConn(conn)
	}
}
