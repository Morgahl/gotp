package gotp

import "fmt"

type Msg interface{}

var _ error = Exit{}

type Exit struct {
	pid    PID
	reason error
}

func NewExit(pid PID, reason error) Exit {
	return Exit{pid, reason}
}

func (e Exit) PID() PID {
	return e.pid
}

func (e Exit) Error() string {
	if e.reason != nil {
		return fmt.Sprintf("%v exit: %v", e.pid, e.reason)
	}
	return fmt.Sprintf("%v exit", e.pid)
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

func NewTimeout(reason error) Timeout {
	return Timeout{reason}
}

func (t Timeout) Error() string {
	return t.reason.Error()
}

func (t Timeout) Is(target error) bool {
	return target == t.reason
}

func (t Timeout) Unwrap() error {
	return t.reason
}
