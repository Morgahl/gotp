package gotp

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/Morgahl/gotp/debug"
)

type Msg interface{}

type MsgMetadata interface {
	Msg
	Metadata() map[Atom]any
}

var _ error = Exit{}

type Exit struct {
	pid    PID
	reason error
}

func NewExit(pid PID, reason error) Exit {
	if reason == nil {
		debug.Throw("NewExit: reason cannot be nil")
	}
	return Exit{pid, reason}
}

func (e Exit) ID() PID {
	return e.pid
}

func (e Exit) Reason() error {
	return e.reason
}

func (e Exit) Error() string {
	if e.reason == nil {
		return fmt.Sprintf("Exit{%s}", e.pid)
	}
	return fmt.Sprintf("Exit{%s, reason: %s}", e.pid, e.reason)
}

func (e Exit) Is(target error) bool {
	return target == e.reason
}

func (e Exit) Unwrap() error {
	return e.reason
}

func (e Exit) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("pid", e.pid),
		slog.Any("reason", e.reason),
	)
}

var _ error = Timeout{}

type Timeout struct {
	reason error
}

func NewTimeout(dur time.Duration) Timeout {
	return Timeout{reason: fmt.Errorf("timeout after %s", dur)}
}

func (t Timeout) Error() string {
	return fmt.Sprintf("Timeout{reason: %s}", t.reason)
}

type Kill struct{}

func (k Kill) Error() string {
	return "Kill{}"
}
