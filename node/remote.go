package node

import (
	"weak"

	"github.com/Morgahl/gotp/process"
)

type Remote struct {
	name    Name
	cookie  Cookie
	address string
	port    uint16
	visible bool
}

func (r Remote) Node() Node {
	return Node{
		Name:   r.name.atom,
		Cookie: r.cookie,
	}
}

type RemoteRef struct {
	pid     process.PID
	nodeRef weak.Pointer[Remote]
}

func newRemoteRef(n *Remote, pid process.PID) RemoteRef {
	return RemoteRef{
		pid:     pid,
		nodeRef: weak.Make(n),
	}
}
