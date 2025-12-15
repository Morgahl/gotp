package gpmd

import (
	"crypto/tls"
	"log/slog"
	"net/rpc"

	"github.com/Morgahl/gotp"
	icrypto "github.com/Morgahl/gotp/internal/crypto"
)

type Client struct {
	cert      tls.Certificate
	rpcClient *rpc.Client
}

func NewClient(cert tls.Certificate) *Client {
	return &Client{
		cert: cert,
	}
}

func (c *Client) Connect() error {
	if c.rpcClient != nil {
		return nil // Already connected
	}

	conn, err := icrypto.Dial(DEFAULT_BIND, c.cert)
	if err != nil {
		return err
	}

	c.rpcClient = rpc.NewClient(conn)
	return nil
}

func (c *Client) Close() error {
	if c.rpcClient == nil {
		return nil // Already closed
	}
	err := c.rpcClient.Close()
	c.rpcClient = nil
	return err
}

func (c *Client) Register(node Node) error {
	slog.Info("Registering node", "node", node)
	err := c.rpcClient.Call("GPMD.Register", node, nil)
	if err != nil {
		slog.Error("Failed to register node", "node", node, "error", err)
	} else {
		slog.Info("Node registered", "node", node)
	}
	return err
}

func (c *Client) Unregister(node Node) error {
	slog.Info("Unregistering node", "node", node)
	return c.rpcClient.Call("GPMD.Unregister", node, nil)
}

func (c *Client) Heartbeat(node Node) error {
	slog.Info("Sending heartbeat for node", "node", node)
	return c.rpcClient.Call("GPMD.Heartbeat", node, nil)
}

func (c *Client) List() (nodes []Node, err error) {
	slog.Info("Listing nodes")
	return nodes, c.rpcClient.Call("GPMD.List", nil, &nodes)
}

func (c *Client) GetByName(name gotp.Atom) (node Node, err error) {
	slog.Info("Getting node by name", "name", name)
	return node, c.rpcClient.Call("GPMD.GetByName", name, &node)
}

func (c *Client) GetByApp(app string) (node Node, err error) {
	slog.Info("Getting node by app", "app", app)
	return node, c.rpcClient.Call("GPMD.GetByApp", app, &node)
}

func (c *Client) GetByHost(host string) (node Node, err error) {
	slog.Info("Getting node by host", "host", host)
	return node, c.rpcClient.Call("GPMD.GetByHost", host, &node)
}
