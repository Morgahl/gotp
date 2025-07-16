package server

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/process"
)

type Supervisable interface {
	ChildSpec() ChildSpec
	StartLink(*process.Process, ...process.SpawnOpt) (Supervised, error)
}

type Supervised interface {
	process.Started
	Supervisable
}

type ChildSpec struct {
	ID gotp.Atom
	Restart
	Shutdown time.Duration
	Type
	Significant bool
	SpawnOpts   []process.SpawnOpt
	Start       func()
}

func (c ChildSpec) String() string {
	return fmt.Sprintf(
		"ChildSpec{ID: %s, Restart: %s, Shutdown: %s, Type: %s, Significant: %t, SpawnOpts: %d}",
		c.ID, c.Restart, c.Shutdown, c.Type, c.Significant, len(c.SpawnOpts))
}

func (c ChildSpec) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("id", string(c.ID)),
		slog.Any("restart", c.Restart),
		slog.Duration("shutdown", c.Shutdown),
		slog.Any("type", c.Type),
		slog.Bool("significant", c.Significant),
		slog.Int("spawnOpts", len(c.SpawnOpts)))
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

func (r Restart) LogValue() slog.Value {
	return slog.StringValue(r.String())
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

func (t Type) LogValue() slog.Value {
	return slog.StringValue(t.String())
}

type AlreadyStarted struct {
	pid process.PID
}

func NewAlreadyStarted(pid process.PID) AlreadyStarted {
	return AlreadyStarted{pid: pid}
}

func (e AlreadyStarted) Error() string {
	return fmt.Sprintf("Child with PID %s is already started", e.pid)
}

func (e AlreadyStarted) LogValue() slog.Value {
	return slog.StringValue(e.Error())
}

type NotStarted struct{}

func NewNotStarted() NotStarted {
	return NotStarted{}
}

func (e NotStarted) Error() string {
	return "not started"
}

func (e NotStarted) LogValue() slog.Value {
	return slog.StringValue(e.Error())
}
