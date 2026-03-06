package node

import (
	"crypto/subtle"
	"crypto/tls"
	"log/slog"
	"net/rpc"

	"github.com/Morgahl/gotp"

	"github.com/Morgahl/gotp/assert"
	icrypto "github.com/Morgahl/gotp/internal/crypto"
)

type Client struct {
	server    Name
	cookie    Cookie
	addr      string
	cert      tls.Certificate
	rpcClient *rpc.Client
}

func NewClient(server gotp.Atom, addr string, cookie Cookie, cert tls.Certificate) *Client {
	assert.GreaterOrEq(len(cookie), 64, "Cookie must be at least 64 bytes long")
	serverName := assert.OkF(ParseName(server))("Invalid server name: %s")

	return &Client{
		server: serverName,
		cookie: cookie,
		addr:   addr,
		cert:   cert,
	}
}

func (c *Client) Connect() {
	slog.Info("Dialing server", "addr", c.addr, "server", c.server)
	conn := assert.OkF(icrypto.Dial(c.addr, c.cert))("Failed to dial server %s: %s", c.addr)

	closeConn := true
	defer func() {
		if closeConn {
			assert.NilF(conn.Close(), "Failed to close connection to %s: %s", c.addr)
		}
	}()

	slog.Info("Connected to server", "addr", c.addr, "server", c.server)

	// HANDSHAKE

	// HandshakeSYN >>>> Server
	slog.Info("Building handshake for server", "addr", c.addr)
	nonce := assert.Ok(icrypto.GenerateNonce())("Failed to generate nonce: %s")
	hmac := icrypto.ComputeHMAC([]byte(c.cookie), nonce)
	n := assert.OkF(conn.Write(nonce[:]))("Failed to write nonce to %s: %s", conn.RemoteAddr())
	assert.EqualF(n, icrypto.NONCE_LENGTH, "Failed to write full nonce to %s: wrote %d bytes, expected %d", conn.RemoteAddr())
	n = assert.OkF(conn.Write(hmac[:]))("Failed to write HMAC to %s: %s", conn.RemoteAddr())
	assert.EqualF(n, icrypto.MAC_LENGTH, "Failed to write full HMAC to %s: wrote %d bytes, expected %d", conn.RemoteAddr())
	slog.Info("Handshake with server sent", "addr", c.addr)

	// HandshakeSYNACK <<<< Server
	rNonce := assert.OkF(icrypto.ReadNonce(conn))("Failed to read rNonce from %s: %s", conn.RemoteAddr())
	rMac := assert.OkF(icrypto.ReadMAC(conn))("Failed to read MAC from %s: %s", conn.RemoteAddr())
	assert.AssertF(icrypto.VerifyMAC([]byte(c.cookie), rNonce, rMac), "Bad handshake invalid HMAC from %s", conn.RemoteAddr())

	// Ensure nonces and HMACs are not equal
	assert.RefuteF(subtle.ConstantTimeCompare(rNonce[:], nonce[:]) == 1, "Bad handshake response nonce cannot match nonce sent to %s", conn.RemoteAddr())
	assert.RefuteF(subtle.ConstantTimeCompare(rMac[:], hmac[:]) == 1, "Bad handshake response HMAC cannot match HMAC sent to %s", conn.RemoteAddr())

	// HandshakeACK >>>> Server
	slog.Info("Building handshake for server", "addr", c.addr)
	rrNonce := assert.Ok(icrypto.GenerateNonce())("Failed to generate nonce: %s")
	rrHmac := icrypto.ComputeHMAC([]byte(c.cookie), rrNonce)
	rrN := assert.OkF(conn.Write(rrNonce[:]))("Failed to write nonce to %s: %s", conn.RemoteAddr())
	assert.EqualF(rrN, icrypto.NONCE_LENGTH, "Failed to write full nonce to %s: wrote %d bytes, expected %d", conn.RemoteAddr())
	rrN = assert.OkF(conn.Write(rrHmac[:]))("Failed to write HMAC to %s: %s", conn.RemoteAddr())
	assert.EqualF(rrN, icrypto.MAC_LENGTH, "Failed to write full HMAC to %s: wrote %d bytes, expected %d", conn.RemoteAddr())
	slog.Info("Handshake with server sent", "addr", c.addr)

	// Confirm >>>> Server
	cN := assert.OkF(conn.Write([]byte{1}))("Failed to write handshake confirmation to %s: %s", conn.RemoteAddr())
	assert.EqualF(cN, 1, "Failed to write handshake confirmation to %s: wrote %d bytes, expected %d", conn.RemoteAddr())

	// Confirm <<<< Server
	var ok [1]byte
	sCN := assert.OkF(conn.Read(ok[:]))("Failed to read handshake confirmation from %s: %s", conn.RemoteAddr())
	assert.EqualF(sCN, 1, "Handshake confirmation failed from %s: expected %d, got %d", conn.RemoteAddr())
	assert.EqualF(ok[0], 1, "Handshake confirmation failed from %s: expected %d, got %d", conn.RemoteAddr())
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
