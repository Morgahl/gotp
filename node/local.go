package node

import (
	"weak"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/debug"
	"github.com/Morgahl/gotp/process"
)

// Local is the RPC server for the local node and is a local singleton. It communicated only with other nodes as a
// peer in a cluster. It will act as a bridge between the local node and the remote nodes for RPC calls.
type Local struct {
	name   Name
	cookie Cookie

	nodeMap map[uint16]*Remote
}

func NewLocal(name gotp.Atom, cookie Cookie) *Local {
	nme, ok := ParseNode(name)
	if ok != nil {
		debug.Throw("name cannot be empty: %s", name)
	}
	if cookie == "" {
		debug.Throw("cookie cannot be empty: %s", cookie)
	}
	return &Local{
		name:    nme,
		cookie:  cookie,
		nodeMap: make(map[uint16]*Remote),
	}
}

func (nl *Local) List(opt gotp.Atom) []gotp.Atom {
	nodes := make([]gotp.Atom, 0, len(nl.nodeMap))
	switch opt {
	case "visible":
		for _, node := range nl.nodeMap {
			if node.visible {
				nodes = append(nodes, node.name.name)
			}
		}

	case "hidden":
		for _, node := range nl.nodeMap {
			if !node.visible {
				nodes = append(nodes, node.name.name)
			}
		}

	case "this":
		nodes = append(nodes, nl.name.name)

	case "all":
		for _, node := range nl.nodeMap {
			nodes = append(nodes, node.name.name)
		}
	}
	return nodes
}

func (nl *Local) remoteRef(pid process.PID) RemoteRef {
	if node, ok := nl.nodeMap[pid.NodeID()]; ok {
		return newRemoteRef(node, pid)
	}
	return RemoteRef{
		pid:     pid,
		nodeRef: weak.Pointer[Remote]{},
	}
}
