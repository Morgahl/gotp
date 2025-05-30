package gotp

import (
	"fmt"
	"time"
)

type Supervisable interface {
	ChildSpec() ChildSpec
	StartLink(PID, time.Duration, ...SpawnOpt) (Supervised, error)
}

type Supervised interface {
	Started
	Supervisable
}

type ChildSpec struct {
	Name string
	Restart
	Shutdown time.Duration
	Type
	Significant bool
	SpawnOpts   []SpawnOpt
}

func (c ChildSpec) String() string {
	return fmt.Sprintf("ChildSpec{Name: %s, Restart: %s, Shutdown: %s, Type: %s, Significant: %t}", c.Name, c.Restart, c.Shutdown, c.Type, c.Significant)
}

type Restart uint8

const (
	PERMANENT Restart = iota
	TRANSIENT
	TEMPORARY
)

func (r Restart) String() string {
	switch r {
	case PERMANENT:
		return "PERMANENT"
	case TRANSIENT:
		return "TRANSIENT"
	case TEMPORARY:
		return "TEMPORARY"
	default:
		return fmt.Sprintf("Restart(%v)", uint8(r))
	}
}

type Type uint8

const (
	WORKER Type = iota
	SUPERVISOR
)

func (t Type) String() string {
	switch t {
	case WORKER:
		return "WORKER"
	case SUPERVISOR:
		return "SUPERVISOR"
	default:
		return fmt.Sprintf("Type(%v)", uint8(t))
	}
}

type AlreadyStarted struct {
	pid PID
}

func NewAlreadyStarted(pid PID) AlreadyStarted {
	return AlreadyStarted{pid: pid}
}

func (e AlreadyStarted) Error() string {
	return fmt.Sprintf("Child with PID %s is already started", e.pid)
}

type NotStarted struct{}

func NewNotStarted() NotStarted {
	return NotStarted{}
}

func (e NotStarted) Error() string {
	return "not started"
}
