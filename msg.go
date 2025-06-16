package gotp

import (
	"fmt"
	"time"
)

type Msg interface{}

type MsgMetadata interface {
	Msg
	Metadata() map[string]any
}

var _ error = Exit{}

type Exit struct {
	pid    PID
	reason error
}

func NewExit(pid PID, reason error) Exit {
	if reason == nil {
		panic("NewExit: reason cannot be nil")
	}
	return Exit{pid, reason}
}

func (e Exit) PID() PID {
	return e.pid
}

func (e Exit) Error() string {
	return fmt.Sprintf("Exit{%v, reason: %v}", e.pid, e.reason)
}

func (e Exit) Is(target error) bool {
	return target == e.reason
}

func (e Exit) Unwrap() error {
	return e.reason
}

var _ error = Timeout{}

type Timeout struct {
	reason error
}

func NewTimeout(dur time.Duration) Timeout {
	return Timeout{reason: fmt.Errorf("timeout after %v", dur)}
}

func (t Timeout) Error() string {
	return fmt.Sprintf("Timeout{reason: %v}", t.reason)
}

type Kill struct{}

func (k Kill) Error() string {
	return "Kill{}"
}
