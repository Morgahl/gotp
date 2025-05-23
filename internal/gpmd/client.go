package gpmd

import (
	"context"
	"encoding/json"
	"net"
	"net/netip"
)

type Client struct {
	ctx  context.Context
	conn net.Conn
	enc  json.Encoder
	dec  json.Decoder
}

func NewClient(ctx context.Context, addr netip.AddrPort) (*Client, error) {
	conn, err := net.Dial("tcp", addr.String())
	if err != nil {
		return nil, err
	}

	return &Client{
		ctx:  ctx,
		conn: conn,
		enc:  *json.NewEncoder(conn),
		dec:  *json.NewDecoder(conn),
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func Do[R JSONable](client *Client, action ActionRequest) (ActionResult[R], error) {
	if err := client.enc.Encode(action); err != nil {
		return ActionResult[R]{}, err
	}

	var result ActionResult[R]
	if err := client.dec.Decode(&result); err != nil {
		return ActionResult[R]{}, err
	}

	return result, nil
}

func Register(client *Client, node Node) (ActionResult[None], error) {
	return Do[None](client, ActionRequest{
		ActionType: ActionRegister,
		Node:       node,
	})
}

func Unregister(client *Client, node Node) (ActionResult[None], error) {
	return Do[None](client, ActionRequest{
		ActionType: ActionUnregister,
		Node:       node,
	})
}

func Heartbeat(client *Client, node Node) (ActionResult[None], error) {
	return Do[None](client, ActionRequest{
		ActionType: ActionHeartbeat,
		Node:       node,
	})
}

func Names(client *Client) (ActionResult[NodeList], error) {
	return Do[NodeList](client, ActionRequest{
		ActionType: ActionNames,
	})
}

func Kill(client *Client) (ActionResult[None], error) {
	return Do[None](client, ActionRequest{
		ActionType: ActionKill,
	})
}
