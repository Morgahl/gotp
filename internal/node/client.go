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

func (c *Client) Connect() error {
	if c.rpcClient != nil {
		return nil // Already connected
	}
	slog.Info("Dialing server", "addr", c.addr, "server", c.server)
	conn, err := icrypto.Dial(c.addr, c.cert)
	if err != nil {
		slog.Error("Failed to dial server", "addr", c.addr, "error", err)
		return err
	}
	slog.Info("Connected to server", "addr", c.addr, "server", c.server)

	// HANDSHAKE
	defer func() {
		if err := debug.Recover(recover(), "Client handshake failed", nil); err != nil {
			slog.Error("Client handshake failed", "error", err)
			debug.AssertNil(conn.Close(), "Failed to close connection to %s: %s", conn.RemoteAddr(), err)
		}
	}()
	slog.Info("Performing handshake with server", "addr", c.addr)
	nonce, err := icrypto.GenerateNonce()
	debug.AssertNil(err, "Failed to generate nonce: %s", err)
	mac := icrypto.ComputeHMAC([]byte(c.cookie), nonce)
	n, err := conn.Write(nonce)
	debug.AssertNil(err, "Failed to write nonce to %s: %s", conn.RemoteAddr(), err)
	debug.Assert(n == icrypto.NONCE_LENGTH, "Failed to write full nonce to %s: wrote %d bytes, expected %d", conn.RemoteAddr(), n, icrypto.NONCE_LENGTH)
	n, err = conn.Write(mac)
	debug.AssertNil(err, "Failed to write HMAC to %s: %s", conn.RemoteAddr(), err)
	debug.Assert(n == icrypto.MAC_LENGTH, "Failed to write full HMAC to %s: wrote %d bytes, expected %d", conn.RemoteAddr(), n, icrypto.MAC_LENGTH)

	slog.Info("Handshake with server completed", "addr", c.addr)
	c.rpcClient = rpc.NewClient(conn)
	slog.Info("RPC client created", "addr", c.addr, "server", c.server)
	return nil
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
