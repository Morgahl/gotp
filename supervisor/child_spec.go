package supervisor

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/process"
)

const (
	DEFAULT_SHUTDOWN = 5 * time.Second
)

type ChildSpec struct {
	ID          gotp.Atom
	Restart     RestartType
	Shutdown    time.Duration
	Type        ChildType
	Significant bool
	Start       func(...process.SpawnOpt) (process.Ref, error)
}

// Equals checks if two ChildSpecs are equal. This cannot and should not be expected to provide uniqueness on the
// `Start` function, as it is not a comparable type. If uniqueness is required, the `ID` field should be used.
func (c *ChildSpec) Equals(other ChildSpec) bool {
	return c.ID == other.ID &&
		c.Restart == other.Restart &&
		c.Shutdown == other.Shutdown &&
		c.Type == other.Type &&
		c.Significant == other.Significant
}

func (c *ChildSpec) ApplyDefaults() {
	if c.Shutdown < 0 {
		c.Shutdown = DEFAULT_SHUTDOWN
	}
}

func (c ChildSpec) String() string {
	return fmt.Sprintf(
		"ChildSpec{ID: %s, Restart: %s, Shutdown: %s, Type: %s, Significant: %t, Start: %t}",
		c.ID, c.Restart, c.Shutdown, c.Type, c.Significant, c.Start != nil)
}

func (c ChildSpec) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("id", c.ID),
		slog.Any("restart", c.Restart),
		slog.Duration("shutdown", c.Shutdown),
		slog.Any("type", c.Type),
		slog.Bool("significant", c.Significant),
		slog.Bool("start", c.Start != nil),
	)
}

func (c *ChildSpec) Apply(opts ...Override) {
	for _, opt := range opts {
		opt(c)
	}
}

type Override func(*ChildSpec)

func ID(id gotp.Atom) Override {
	return func(spec *ChildSpec) {
		spec.ID = id
	}
}

func Restart(restart RestartType) Override {
	return func(spec *ChildSpec) {
		spec.Restart = restart
	}
}

func Shutdown(shutdown time.Duration) Override {
	return func(spec *ChildSpec) {
		spec.Shutdown = shutdown
	}
}

func Type(t ChildType) Override {
	return func(spec *ChildSpec) {
		spec.Type = t
	}
}

func Significant(significant bool) Override {
	return func(spec *ChildSpec) {
		spec.Significant = significant
	}
}

func Start(start func(...process.SpawnOpt) (process.Ref, error)) Override {
	return func(spec *ChildSpec) {
		spec.Start = start
	}
}

type RestartType uint8

const (
	PERMANENT RestartType = iota
	TRANSIENT
	TEMPORARY
)

func (r RestartType) String() string {
	switch r {
	case PERMANENT:
		return "PERMANENT"
	case TRANSIENT:
		return "TRANSIENT"
	case TEMPORARY:
		return "TEMPORARY"
	}
	return fmt.Sprintf("RestartType(%v)", uint8(r))
}

func (r RestartType) LogValue() slog.Value {
	return slog.StringValue(r.String())
}

type ChildType uint8

const (
	WORKER ChildType = iota
	SUPERVISOR
)

func (t ChildType) String() string {
	switch t {
	case WORKER:
		return "WORKER"
	case SUPERVISOR:
		return "SUPERVISOR"
	default:
		return fmt.Sprintf("Type(%v)", uint8(t))
	}
}

func (t ChildType) LogValue() slog.Value {
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
