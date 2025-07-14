package supervisor

import (
	"fmt"
	"log/slog"

	"github.com/Morgahl/gotp/process"
	"github.com/Morgahl/gotp/server"
)

type startChild struct {
	child server.Supervisable
}

func (s startChild) String() string {
	return fmt.Sprintf("startChild{spec: %s}", s.child.ChildSpec())
}

func (s startChild) LogValue() slog.Value {
	return slog.StringValue(s.String())
}

type stopChild struct {
	pid process.PID
}

func (s stopChild) String() string {
	return fmt.Sprintf("stopChild{pid: %s}", s.pid)
}

func (s stopChild) LogValue() slog.Value {
	return slog.StringValue(s.String())
}

type InvalidChild struct {
	child server.Supervisable
}

func NewInvalidChild(child server.Supervisable) InvalidChild {
	return InvalidChild{child: child}
}

func (e InvalidChild) Error() string {
	return fmt.Sprintf("invalid child: %T", e.child)
}
