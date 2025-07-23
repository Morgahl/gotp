package node

import (
	"github.com/Morgahl/gotp"
)

type Local struct {
	Node
	started bool
}

func newLocal(sname gotp.Atom, cookie Cookie) *Local {
	return &Local{
		Node: Node{
			Name:   sname,
			Cookie: cookie,
		},
	}
}

func (l *Local) Alive(_ any, alive *bool) error {
	*alive = l.started
	return nil
}

// func (l *Local)
