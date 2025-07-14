package gpmd

import (
	"encoding/json"
	"fmt"
	"net/netip"
	"strings"
)

type Node struct {
	Name string         `json:"name"`
	Host netip.AddrPort `json:"host"`
}

func NodeFromString(s string) (Node, error) {
	parts := strings.Split(s, "@")
	if len(parts) != 2 {
		return Node{}, fmt.Errorf("invalid node format; expected name@host")
	}
	return NewNode(parts[0], parts[1])
}

func NewNode(name, host string) (Node, error) {
	addr, err := netip.ParseAddrPort(host)
	if err != nil {
		return Node{}, fmt.Errorf("failed to parse host %q: %w", host, err)
	}

	return Node{
		Name: name,
		Host: addr,
	}, nil
}

func (g Node) String() string {
	return fmt.Sprintf("%s@%s", g.Name, g.Host)
}

type NodeList []Node

func (g NodeList) MarshalJSON() ([]byte, error) {
	return json.Marshal(g)
}

func (g NodeList) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &g)
}
