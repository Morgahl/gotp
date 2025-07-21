package node

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/rpc"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/debug"

	icrypto "github.com/Morgahl/gotp/internal/crypto"
	gpmd "github.com/Morgahl/gotp/internal/gpmd_2"
)

type Server struct {
	name       Name
	cookie     gotp.Atom
	local      *Local
	rpcServer  *rpc.Server
	cert       tls.Certificate
	gpmdClient *gpmd.Client
	listener   net.Listener
	hidden     map[gotp.Atom]*Client
	visible    map[gotp.Atom]*Client
}

func NewServer(sname gotp.Atom, cookie gotp.Atom) *Server {
	debug.Assert(len(cookie) >= 64, "Cookie must be at least 64 bytes long")
	name, err := ParseName(sname)
	debug.AssertNil(err, "Invalid node name: %s", err)
	l := newLocal(name.Name, cookie)
	rpcServer := rpc.NewServer()
	debug.AssertNil(rpcServer.Register(l), "Failed to register local node")

	cert, err := icrypto.GenerateSelfSignedCert(name.App, cookie)
	debug.AssertNil(err, "Failed to generate self-signed certificate: %s", err)

	s := &Server{
		name:      name,
		cookie:    cookie,
		local:     l,
		cert:      cert,
		rpcServer: rpcServer,
		hidden:    make(map[gotp.Atom]*Client),
		visible:   make(map[gotp.Atom]*Client),
	}

	return s
}

func (s *Server) Start() {
	debug.AssertNil(s.listener, "Server already started")
	slog.Info("Starting node server", "sname", s.name)
	s.gpmdClient = gpmd.NewClient(s.cert)
	slog.Info("Connecting to GPMD", "bind", gpmd.DEFAULT_BIND)
	err := s.gpmdClient.Connect()
	slog.Info("Connected to GPMD", "err", err)
	debug.AssertNil(err, "Failed to connect to GPMD: %s", err)
	debug.AssertNil(s.listener, "Server already started")
	s.listener, err = icrypto.Listen(s.name.Host, s.cert)
	debug.AssertNil(err, "Failed to start listener: %s", err)
	slog.Info("Node server started", "sname", s.name.App, "bind", s.listener.Addr().String())
	n, err := gpmd.NewNode(s.name.Name, string(s.name.App), s.listener.Addr().String())
	debug.AssertNil(err, "Failed to create GPMD node: %s", err)
	slog.Info("Registering with GPMD", "node.name", n.Name, "node.host", n.Host)
	err = s.gpmdClient.Register(n)
	slog.Info("Registered with GPMD", "node", n, "err", err)
	debug.AssertNil(err, "Failed to register with GPMD: %s", err)
	go s.listen()
}

func (s *Server) Stop() {
	debug.AssertNotNil(s.listener, "Server not started")
	debug.AssertNil(s.gpmdClient.Close(), "Failed to close GPMD client")
	s.gpmdClient = nil
	debug.AssertNil(s.listener.Close(), "Failed to close listener")
	slog.Info("Stopping node server", "sname", s.name)
	s.listener = nil
}

func (s *Server) Connect(server gotp.Atom, cookie gotp.Atom, visible bool) error {
	debug.AssertNotEqual(server, s.name.Name, "Cannot connect to self")
	name, err := ParseName(server)
	if err != nil {
		return fmt.Errorf("invalid node name %s: %w", server, err)
	}
	node, err := s.gpmdClient.GetByName(name.Name)
	if err != nil {
		slog.Error("Failed to get node by name", "name", name.Name, "error", err)
		return err
	}
	slog.Info("Building client for node", "node", node)
	client := NewClient(server, node.Host, cookie, s.cert)
	slog.Info("Connecting to node", "node", node)
	client.Connect()
	slog.Info("Connected to node", "node", node, "server", server)

	if visible {
		if _, ok := s.visible[server]; ok {
			return fmt.Errorf("already connected to visible node %s", server)
		}
		s.visible[server] = client
		slog.Info("Added visible node", "node", node, "server", server)
	} else {
		if _, ok := s.hidden[server]; ok {
			return fmt.Errorf("already connected to hidden node %s", server)
		}
		s.hidden[server] = client
		slog.Info("Added hidden node", "node", node, "server", server)
	}
	slog.Info("Node connected", "node", node, "server", server)
	return nil
}

func (s *Server) Disconnect(server gotp.Atom) error {
	if client, ok := s.hidden[server]; ok {
		delete(s.hidden, server)
		return client.Close()
	}
	if client, ok := s.visible[server]; ok {
		delete(s.visible, server)
		return client.Close()
	}
	return fmt.Errorf("not connected to %s", server)
}

func (s *Server) ListVisible() []gotp.Atom {
	nodes := make([]gotp.Atom, 0, len(s.visible))
	for name := range s.visible {
		nodes = append(nodes, name)
	}
	return nodes
}

func (s *Server) ListHidden() []gotp.Atom {
	nodes := make([]gotp.Atom, 0, len(s.hidden))
	for name := range s.hidden {
		nodes = append(nodes, name)
	}
	return nodes
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		debug.AssertNil(err, "Failed to accept connection from %s: %s", conn.RemoteAddr(), err)
		// TODO: FUTURE HOME OF A LINKED PROCESS SPAWN INSTEAD OF THIS GO ROUTINE
		go func(c net.Conn, cookie Cookie) {
			defer c.Close()
			// HANDSHAKE

			// Handshake <<<< Client
			nonce, err := icrypto.ReadNonce(c)
			debug.AssertNil(err, "Failed to read nonce from %s: %s", c.RemoteAddr(), err)
			hmac, err := icrypto.ReadHMAC(c)
			debug.AssertNil(err, "Failed to read HMAC from %s: %s", c.RemoteAddr(), err)
			debug.Assert(icrypto.VerifyMAC([]byte(cookie), nonce, hmac), "Bad handshake invalid HMAC from %s", c.RemoteAddr())

			// Handshake >>>> Client
			slog.Info("Building handshake response for client", "addr", c.RemoteAddr())
			rNonce, err := icrypto.GenerateNonce()
			debug.AssertNil(err, "Failed to generate nonce: %s", err)
			rHmac := icrypto.ComputeHMAC([]byte(s.cookie), rNonce)
			n, err := conn.Write(rNonce[:])
			debug.AssertNil(err, "Failed to write nonce to %s: %s", conn.RemoteAddr(), err)
			debug.Assert(n == icrypto.NONCE_LENGTH, "Failed to write full nonce to %s: wrote %d bytes, expected %d", conn.RemoteAddr(), n, icrypto.NONCE_LENGTH)
			n, err = conn.Write(rHmac[:])
			debug.AssertNil(err, "Failed to write HMAC to %s: %s", conn.RemoteAddr(), err)
			debug.Assert(n == icrypto.MAC_LENGTH, "Failed to write full HMAC to %s: wrote %d bytes, expected %d", conn.RemoteAddr(), n, icrypto.MAC_LENGTH)
			slog.Info("Handshake with client sent", "addr", conn.RemoteAddr())

			// Confirm <<<< Client
			var ok [1]byte
			_, err = conn.Read(ok[:])
			debug.AssertNil(err, "Failed to read handshake confirmation from %s: %s", conn.RemoteAddr(), err)
			debug.Assert(ok[0] == 1, "Handshake confirmation failed from %s: expected 1, got %d", conn.RemoteAddr(), ok[0])

			// Confirm >>>> Client
			_, err = conn.Write([]byte{1})
			debug.AssertNil(err, "Failed to write handshake confirmation to %s: %s", conn.RemoteAddr(), err)
			slog.Info("Handshake with client confirmed", "addr", conn.RemoteAddr())

			// Create RPC server
			s.rpcServer.ServeConn(c)
		}(conn, s.local.Cookie)
	}
}
