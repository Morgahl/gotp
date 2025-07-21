package node

import (
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
	slog.Info("Connected to server", "addr", c.addr, "server", c.server)

	// HANDSHAKE

	// Handshake >>>> Server
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

	// Handshake <<<< Server
	rNonce, err := icrypto.ReadNonce(conn)
	debug.AssertNil(err, "Failed to read rNonce from %s: %s", conn.RemoteAddr(), err)
	rHmac, err := icrypto.ReadHMAC(conn)
	debug.AssertNil(err, "Failed to read HMAC from %s: %s", conn.RemoteAddr(), err)
	debug.Assert(icrypto.VerifyMAC([]byte(c.cookie), rNonce, rHmac), "Bad handshake invalid HMAC from %s", conn.RemoteAddr())

	// Confirm >>>> Server
	_, err = conn.Write([]byte{1})
	debug.AssertNil(err, "Failed to write handshake confirmation to %s: %s", conn.RemoteAddr(), err)

	// Confirm <<<< Server
	var ok [1]byte
	_, err = conn.Read(ok[:])
	debug.AssertNil(err, "Failed to read handshake confirmation from %s: %s", conn.RemoteAddr(), err)
	debug.Assert(ok[0] == 1, "Handshake confirmation failed from %s: expected 1, got %d", conn.RemoteAddr(), ok[0])
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
