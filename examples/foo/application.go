package foo

import (
	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/application"
	"github.com/Morgahl/gotp/supervisor"
)

type FooApplication struct{}

func (FooApplication) Name() string {
	return "FooApp"
}

func (FooApplication) Version() application.Version {
	return application.Ver(0, 1, 0, "alpha")
}

func (FooApplication) Start(st application.StartType) (gotp.Supervisable, error) {
	srvr0 := NewServer("alice")
	srvr1 := NewServer("bob")
	srvr2 := NewServer("charlie")
	static := NewSupervisor("foo", supervisor.Flags{}, srvr0, srvr1, srvr2)
	srvr3 := NewServer("dave")
	srvr4 := NewServer("eve")
	srvr5 := NewServer("frank")
	root := NewSupervisor("bar", supervisor.Flags{}, static, srvr3, srvr4, srvr5)
	return root, nil
}
