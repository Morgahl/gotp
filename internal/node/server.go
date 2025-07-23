package node

import (
	"crypto/subtle"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/rpc"

	"github.com/Morgahl/gotp"

	"github.com/Morgahl/gotp/assert"
	icrypto "github.com/Morgahl/gotp/internal/crypto"
	gpmd "github.com/Morgahl/gotp/internal/gpmd_2"
)

var (
	local *Server
)

type Server struct {
	name       Name
	cookie     Cookie
	rpcServer  *rpc.Server
	cert       tls.Certificate
	gpmdClient *gpmd.Client
	listener   net.Listener
	known      map[gotp.Atom]Cookie
	hidden     map[gotp.Atom]*Client
	visible    map[gotp.Atom]*Client
}

func NewServer(sname gotp.Atom, cookie Cookie) *Server {
	assert.NilPtr(local, "Server already initialized")
	assert.GreaterOrEq(len(cookie), 64, "Cookie must be at least 64 bytes long")
	assert.NotZero(sname, "Server name cannot be empty")
	name := assert.Ok(ParseName(sname))("Invalid node name: %s")
	l := newLocal(name.Name, cookie)
	rpcServer := rpc.NewServer()
	assert.NilF(rpcServer.Register(l), "Failed to register local node: %s")
	cert := assert.OkF(icrypto.GenerateSelfSignedCert(name.App, gotp.Atom(cookie)))("Failed to generate self-signed certificate: %s")

	local = &Server{
		name:      name,
		cookie:    cookie,
		cert:      cert,
		rpcServer: rpcServer,
		known:     make(map[gotp.Atom]Cookie),
		hidden:    make(map[gotp.Atom]*Client),
		visible:   make(map[gotp.Atom]*Client),
	}

	return local
}

func (s *Server) Start() {
	assert.Nil(s.listener, "Server already started")
	slog.Info("Starting node server", "sname", s.name)
	s.gpmdClient = gpmd.NewClient(s.cert)
	slog.Info("Connecting to GPMD", "bind", gpmd.DEFAULT_BIND)
	assert.NilF(s.gpmdClient.Connect(), "Failed to connect to GPMD: %s")
	assert.Nil(s.listener, "Server already started")
	s.listener = assert.OkF(icrypto.Listen(s.name.Host, s.cert))("Failed to start listener: %s")
	slog.Info("Node server started", "sname", s.name.App, "bind", s.listener.Addr().String())
	n := assert.OkF(gpmd.NewNode(s.name.Name, s.name.App, s.listener.Addr().String()))("Failed to create GPMD node: %s")
	slog.Info("Registering with GPMD", "node.name", n.Name, "node.host", n.Host)
	assert.NilF(s.gpmdClient.Register(n), "Failed to register with GPMD: %s")
	slog.Info("Registered with GPMD", "node", n)
	go s.listen()
}

func (s *Server) Stop() {
	assert.NotNil(s.listener, "Server not started")
	assert.Nil(s.gpmdClient.Close(), "Failed to close GPMD client")
	s.gpmdClient = nil
	assert.Nil(s.listener.Close(), "Failed to close listener")
	slog.Info("Stopping node server", "sname", s.name)
	s.listener = nil
}

func (s *Server) Connect(server gotp.Atom, cookie Cookie, visible bool) error {
	assert.NotEqual(server, s.name.Name, "Cannot connect to self")
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
	s.known[server] = cookie
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

func (s *Server) List(name gotp.Atom) ([]gotp.Atom, error) {
	switch name {
	case "", "visible":
		nodes := make([]gotp.Atom, 0, len(s.visible))
		for name := range s.visible {
			nodes = append(nodes, name)
		}
		return nodes, nil
	case "hidden":
		nodes := make([]gotp.Atom, 0, len(s.hidden))
		for name := range s.hidden {
			nodes = append(nodes, name)
		}
		return nodes, nil
	case "this":
		return []gotp.Atom{s.name.Name}, nil
	case "connected":
		nodes := make([]gotp.Atom, 0, len(s.visible)+len(s.hidden)+1)
		for name := range s.visible {
			nodes = append(nodes, name)
		}
		for name := range s.hidden {
			nodes = append(nodes, name)
		}
		nodes = append(nodes, s.name.Name)
		return nodes, nil
	}
	return nil, fmt.Errorf("unknown list type %s", name)
}

func (s *Server) listen() {
	for {
		conn := assert.OkF(s.listener.Accept())("Failed to accept connection on %s: %s", s.listener.Addr())
		// TODO: FUTURE HOME OF A LINKED PROCESS SPAWN INSTEAD OF THIS GO ROUTINE
		go func(c net.Conn, cookie Cookie) {
			defer c.Close()

			// HANDSHAKE

			// HandshakeSYN <<<< Client
			nonce := assert.OkF(icrypto.ReadNonce(c))("Failed to read nonce from %s: %s", c.RemoteAddr())
			hmac := assert.OkF(icrypto.ReadHMAC(c))("Failed to read HMAC from %s: %s", c.RemoteAddr())
			assert.AssertF(icrypto.VerifyMAC([]byte(cookie), nonce, hmac), "Bad handshake invalid HMAC from %s", c.RemoteAddr())

			// HandshakeSYNACK >>>> Client
			slog.Info("Building handshake response for client", "addr", c.RemoteAddr())
			rNonce := assert.OkF(icrypto.GenerateNonce())("Failed to generate nonce: %s")
			rHmac := icrypto.ComputeHMAC([]byte(cookie), rNonce)
			rN := assert.OkF(c.Write(rNonce[:]))("Failed to write nonce to %s: %s", c.RemoteAddr())
			assert.EqualF(rN, icrypto.NONCE_LENGTH, "Failed to write full nonce to %s: wrote %d bytes, expected %d", c.RemoteAddr())
			rH := assert.OkF(c.Write(rHmac[:]))("Failed to write HMAC to %s: %s", c.RemoteAddr())
			assert.EqualF(rH, icrypto.MAC_LENGTH, "Failed to write full HMAC to %s: wrote %d bytes, expected %d", c.RemoteAddr())
			slog.Info("Handshake with client sent", "addr", conn.RemoteAddr())

			// HandshakeACK <<<< Client
			rrNonce := assert.OkF(icrypto.ReadNonce(c))("Failed to read nonce from %s: %s", c.RemoteAddr())
			rrHmac := assert.OkF(icrypto.ReadHMAC(c))("Failed to read HMAC from %s: %s", c.RemoteAddr())
			assert.AssertF(icrypto.VerifyMAC([]byte(cookie), rrNonce, rrHmac), "Bad handshake invalid HMAC from %s", c.RemoteAddr())

			// Ensure nonces and HMACs are not equal
			assert.RefuteF(subtle.ConstantTimeCompare(nonce[:], rNonce[:]) == 1, "Bad handshake response nonce cannot match nonce sent to %s", conn.RemoteAddr())
			assert.RefuteF(subtle.ConstantTimeCompare(nonce[:], rrNonce[:]) == 1, "Bad handshake response nonce cannot match nonce sent to %s", conn.RemoteAddr())
			assert.RefuteF(subtle.ConstantTimeCompare(rNonce[:], rrNonce[:]) == 1, "Bad handshake response nonce cannot match nonce sent to %s", conn.RemoteAddr())
			assert.RefuteF(subtle.ConstantTimeCompare(hmac[:], rHmac[:]) == 1, "Bad handshake response HMAC cannot match HMAC sent to %s", conn.RemoteAddr())
			assert.RefuteF(subtle.ConstantTimeCompare(hmac[:], rrHmac[:]) == 1, "Bad handshake response HMAC cannot match HMAC sent to %s", conn.RemoteAddr())
			assert.RefuteF(subtle.ConstantTimeCompare(rHmac[:], rrHmac[:]) == 1, "Bad handshake response HMAC cannot match HMAC sent to %s", conn.RemoteAddr())

			// Confirm >>>> Client
			cN := assert.OkF(c.Write([]byte{1}))("Failed to write handshake confirmation to %s: %s", conn.RemoteAddr())
			assert.EqualF(cN, 1, "Failed to write handshake confirmation to %s: wrote %d bytes, expected %d", conn.RemoteAddr())

			// Confirm <<<< Client
			var ok [1]byte
			sCN := assert.OkF(c.Read(ok[:]))("Failed to read handshake confirmation from %s: %s", conn.RemoteAddr())
			assert.EqualF(sCN, 1, "Handshake confirmation failed from %s: expected %d, got %d", conn.RemoteAddr())
			assert.EqualF(ok[0], 1, "Handshake confirmation failed from %s: expected %d, got %d", conn.RemoteAddr())
			slog.Info("Handshake with client confirmed", "addr", c.RemoteAddr())

			// Create RPC server
			s.rpcServer.ServeConn(c)
		}(conn, s.cookie)
	}
}
