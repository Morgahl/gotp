package node

import (
	"github.com/Morgahl/gotp"
)

type Node struct {
	Name   gotp.Atom
	Cookie Cookie
}

type Cookie gotp.Atom

func (c Cookie) String() string {
	return gotp.Atom(c).String()
}
