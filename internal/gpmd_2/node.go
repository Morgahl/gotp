package gpmd

import (
	"encoding/json"
	"fmt"
	"net/netip"

	"github.com/Morgahl/gotp"
)

type Node struct {
	Name gotp.Atom
	App  gotp.Atom
	Host string
}

func NewNode(name gotp.Atom, app gotp.Atom, host string) (Node, error) {
	addr, err := netip.ParseAddrPort(host)
	if err != nil {
		return Node{}, fmt.Errorf("failed to parse host %q: %w", host, err)
	}

	return Node{
		Name: name,
		App:  app,
		Host: addr.String(),
	}, nil
}

func (g Node) String() string {
	return fmt.Sprintf("%s@%s", g.App, g.Host)
}

type NodeList []Node

func (g NodeList) MarshalJSON() ([]byte, error) {
	return json.Marshal(g)
}

func (g NodeList) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &g)
}
