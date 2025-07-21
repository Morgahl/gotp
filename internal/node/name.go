package node

import (
	"errors"
	"strings"

	"github.com/Morgahl/gotp"
)

type Name struct {
	Name gotp.Atom
	App  gotp.Atom
	Host string
}

func ParseName(name gotp.Atom) (Name, error) {
	parts := strings.Split(name.String(), "@")
	if len(parts) != 2 {
		return Name{}, errors.New("invalid node name, must have exactly one '@' character: " + name.String())
	}
	if parts[0] == "" || parts[1] == "" {
		return Name{}, errors.New("invalid node name, must have non-empty name and host: " + name.String())
	}
	return Name{
		Name: name,
		App:  gotp.Atom(parts[0]),
		Host: parts[1],
	}, nil
}

func (n Name) String() string {
	return n.App.String() + "@" + n.Host
}

func (n Name) HostType() string {
	if strings.Contains(n.Host, ":") {
		return "ipv6"
	}
	return "ipv4"
}
