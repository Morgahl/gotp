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
	return NewSupervisor("foo", supervisor.Flags{},
		NewSupervisor("bar", supervisor.Flags{},
			NewServer("alice"),
			NewServer("bob"),
			NewServer("charlie")),
		NewServer("dave"),
		NewServer("eve"),
		NewServer("frank"),
	), nil
}
