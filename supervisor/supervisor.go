package supervisor

import (
	"fmt"
	"time"

	"github.com/Morgahl/gotp"
)

type Supervisor interface {
	gotp.Supervisable
	StartChild(gotp.Supervisable, time.Duration) error
	StopChild(gotp.PID, time.Duration) error
}

type Strategy uint8

const (
	ONE_FOR_ONE Strategy = iota
	ONE_FOR_ALL
	REST_FOR_ONE
)

func (s Strategy) String() string {
	switch s {
	case ONE_FOR_ONE:
		return "ONE_FOR_ONE"
	case ONE_FOR_ALL:
		return "ONE_FOR_ALL"
	case REST_FOR_ONE:
		return "REST_FOR_ONE"
	default:
		return fmt.Sprintf("Strategy(%v)", uint8(s))
	}
}

type AutoShutdown uint8

const (
	NEVER AutoShutdown = iota
	ANY_SIGNIFICANT
	ALL_SIGNIFICANT
)

func (a AutoShutdown) String() string {
	switch a {
	case NEVER:
		return "NEVER"
	case ANY_SIGNIFICANT:
		return "ANY_SIGNIFICANT"
	case ALL_SIGNIFICANT:
		return "ALL_SIGNIFICANT"
	default:
		return fmt.Sprintf("AutoShutdown(%v)", uint8(a))
	}
}

type Flags struct {
	AutoShutdown
	MaxRestarts uint
	ResetPeriod time.Duration
	Shutdown    time.Duration
	Strategy
}

func (f Flags) String() string {
	return fmt.Sprintf("Flags{AutoShutdown: %s, MaxRestarts: %d, ResetPeriod: %s, Shutdown: %s, Strategy: %s}", f.AutoShutdown, f.MaxRestarts, f.ResetPeriod, f.Shutdown, f.Strategy)
}

func (f Flags) ApplyDefaults() Flags {
	if f.MaxRestarts == 0 {
		f.MaxRestarts = 3
	}
	if f.ResetPeriod == 0 {
		f.ResetPeriod = 5 * time.Second
	}
	if f.Shutdown == 0 {
		f.Shutdown = 30 * time.Second
	}
	return f
}

func StartChild(s Supervisor, child gotp.Supervisable, timeout time.Duration) error {
	return s.StartChild(child, timeout)
}

func StopChild(s Supervisor, pid gotp.PID, timeout time.Duration) error {
	if pid == gotp.PIDZero() {
		// better error
		return fmt.Errorf("cannot stop child with zero PID")
	}
	return s.StopChild(pid, timeout)
}

type child struct {
	supervisable gotp.Supervisable
	running      gotp.Running
	restart      restart
}

type restart struct {
	count uint
	at    time.Time
}
