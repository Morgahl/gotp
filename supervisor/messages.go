package supervisor

import (
	"fmt"

	"github.com/Morgahl/gotp"
)

type startChild struct {
	child gotp.Supervisable
}

type stopChild struct {
	pid gotp.PID
}

type InvalidChild struct {
	child gotp.Supervisable
}

func NewInvalidChild(child gotp.Supervisable) InvalidChild {
	return InvalidChild{child: child}
}

func (e InvalidChild) Error() string {
	return fmt.Sprintf("invalid child: %T", e.child)
}
