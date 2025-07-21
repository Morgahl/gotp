package node

import (
	"crypto/subtle"
	"crypto/tls"
	"log/slog"
	"net/rpc"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/debug"

	icrypto "github.com/Morgahl/gotp/internal/crypto"
)

type Client struct {
	server    Name
	cookie    Cookie
	addr      string
	cert      tls.Certificate
	rpcClient *rpc.Client
}

func NewClient(server gotp.Atom, addr string, cookie gotp.Atom, cert tls.Certificate) *Client {
	debug.Assert(len(cookie) >= 64, "Cookie must be at least 64 bytes long")
	serverName, err := ParseName(server)
	debug.AssertNil(err, "Invalid server name: %s", err)

	return &Client{
		server: serverName,
		cookie: Cookie(cookie),
		addr:   addr,
		cert:   cert,
	}
}

func (c *Client) Connect() {
	slog.Info("Dialing server", "addr", c.addr, "server", c.server)
	conn, err := icrypto.Dial(c.addr, c.cert)
	debug.AssertNil(err, "Failed to dial server %s: %s", c.addr, err)

	closeConn := true
	defer func() {
		if closeConn {
			debug.AssertNil(conn.Close(), "Failed to close connection to %s: %s", c.addr, err)
		}
	}()

	slog.Info("Connected to server", "addr", c.addr, "server", c.server)

	// HANDSHAKE

	// HandshakeSYN >>>> Server
	slog.Info("Building handshake for server", "addr", c.addr)
	nonce, err := icrypto.GenerateNonce()
	debug.AssertNil(err, "Failed to generate nonce: %s", err)
	hmac := icrypto.ComputeHMAC([]byte(c.cookie), nonce)
	n, err := conn.Write(nonce[:])
	debug.AssertNil(err, "Failed to write nonce to %s: %s", conn.RemoteAddr(), err)
	debug.Assert(n == icrypto.NONCE_LENGTH, "Failed to write full nonce to %s: wrote %d bytes, expected %d", conn.RemoteAddr(), n, icrypto.NONCE_LENGTH)
	n, err = conn.Write(hmac[:])
	debug.AssertNil(err, "Failed to write HMAC to %s: %s", conn.RemoteAddr(), err)
	debug.Assert(n == icrypto.MAC_LENGTH, "Failed to write full HMAC to %s: wrote %d bytes, expected %d", conn.RemoteAddr(), n, icrypto.MAC_LENGTH)
	slog.Info("Handshake with server sent", "addr", c.addr)

	// HandshakeSYNACK <<<< Server
	rNonce, err := icrypto.ReadNonce(conn)
	debug.AssertNil(err, "Failed to read rNonce from %s: %s", conn.RemoteAddr(), err)
	rHmac, err := icrypto.ReadHMAC(conn)
	debug.AssertNil(err, "Failed to read HMAC from %s: %s", conn.RemoteAddr(), err)
	debug.Assert(icrypto.VerifyMAC([]byte(c.cookie), rNonce, rHmac), "Bad handshake invalid HMAC from %s", conn.RemoteAddr())

	// Ensure nonces and HMACs are not equal
	debug.Refute(subtle.ConstantTimeCompare(rNonce[:], nonce[:]) == 1, "Bad handshake response nonce cannot match nonce sent to %s", conn.RemoteAddr())
	debug.Refute(subtle.ConstantTimeCompare(rHmac[:], hmac[:]) == 1, "Bad handshake response HMAC cannot match HMAC sent to %s", conn.RemoteAddr())

	// HandshakeACK >>>> Server
	slog.Info("Building handshake for server", "addr", c.addr)
	rrNonce, err := icrypto.GenerateNonce()
	debug.AssertNil(err, "Failed to generate nonce: %s", err)
	rrHmac := icrypto.ComputeHMAC([]byte(c.cookie), rrNonce)
	rrN, err := conn.Write(rrNonce[:])
	debug.AssertNil(err, "Failed to write nonce to %s: %s", conn.RemoteAddr(), err)
	debug.Assert(rrN == icrypto.NONCE_LENGTH, "Failed to write full nonce to %s: wrote %d bytes, expected %d", conn.RemoteAddr(), n, icrypto.NONCE_LENGTH)
	rrN, err = conn.Write(rrHmac[:])
	debug.AssertNil(err, "Failed to write HMAC to %s: %s", conn.RemoteAddr(), err)
	debug.Assert(rrN == icrypto.MAC_LENGTH, "Failed to write full HMAC to %s: wrote %d bytes, expected %d", conn.RemoteAddr(), n, icrypto.MAC_LENGTH)
	slog.Info("Handshake with server sent", "addr", c.addr)

	// Confirm >>>> Server
	_, err = conn.Write([]byte{1})
	debug.AssertNil(err, "Failed to write handshake confirmation to %s: %s", conn.RemoteAddr(), err)

	// Confirm <<<< Server
	var ok [1]byte
	_, err = conn.Read(ok[:])
	debug.AssertNil(err, "Failed to read handshake confirmation from %s: %s", conn.RemoteAddr(), err)
	debug.Assert(ok[0] == 1, "Handshake confirmation failed from %s: expected 1, got %d", conn.RemoteAddr(), ok[0])
	closeConn = false
	slog.Info("Handshake with server confirmed", "addr", c.addr)

	// Create RPC client
	c.rpcClient = rpc.NewClient(conn)
	slog.Info("RPC client created", "addr", c.addr, "server", c.server)
}

func (c *Client) Close() error {
	if c.rpcClient == nil {
		return nil
	}
	err := c.rpcClient.Close()
	c.rpcClient = nil
	return err
}

func (c *Client) Multiply(args Args) (reply int, err error) {
	debug.AssertNotNil(c.rpcClient, "call Connect() before calling RPC methods")
	err = c.rpcClient.Call("Local.Multiply", &args, &reply)
	return reply, err
}

func (c *Client) Divide(args Args) (quo Quotient, err error) {
	debug.AssertNotNil(c.rpcClient, "call Connect() before calling RPC methods")
	err = c.rpcClient.Call("Local.Divide", &args, &quo)
	return quo, err
}
