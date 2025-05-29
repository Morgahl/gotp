package application

import (
	"fmt"

	"github.com/Morgahl/gotp"
)

type Application interface {
	Name() string
	Version() Version
	Start(StartType) (gotp.Supervisable, error)
}

type PrepareStop interface {
	Application
	PrepareStop() error
}

type Stop interface {
	Application
	Stop() error
}

type startType uint8

const (
	NORMAL startType = iota
	TAKEOVER
	FAILOVER
)

type StartType struct {
	startType startType
	node      gotp.Node
}

func Normal() StartType {
	return StartType{startType: NORMAL}
}

func Takeover(node gotp.Node) StartType {
	return StartType{startType: TAKEOVER, node: node}
}

func Failover(node gotp.Node) StartType {
	return StartType{startType: FAILOVER, node: node}
}

func (st StartType) String() string {
	switch st.startType {
	case NORMAL:
		return "normal"
	case TAKEOVER:
		return fmt.Sprintf("takeover=%s", st.node)
	case FAILOVER:
		return fmt.Sprintf("failover=%s", st.node)
	default:
		return "unknown"
	}
}

func (st StartType) IsNormal() bool {
	return st.startType == NORMAL
}

func (st StartType) IsTakeover() (node gotp.Node, ok bool) {
	if st.startType == TAKEOVER {
		return st.node, true
	}
	return
}

func (st StartType) IsFailover() (node gotp.Node, ok bool) {
	if st.startType == FAILOVER {
		return st.node, true
	}
	return
}
