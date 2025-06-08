package foo

import (
	"errors"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/application"
	"github.com/Morgahl/gotp/supervisor"
)

type FooApplication struct{}

func (FooApplication) Name() string {
	return "FooApplication"
}

func (FooApplication) Version() application.Version {
	return application.Ver(0, 1, 0, "alpha")
}

func (FooApplication) Start(st application.StartType) (gotp.Supervisable, error) {
	if st.IsNormal() {
		return NewFooSupervisor("foo", supervisor.Flags{},
			NewFooSupervisor("bar", supervisor.Flags{},
				NewFooServer("alice"),
				NewFooServer("bob"),
				NewFooServer("charlie")),
			NewFooServer("dave"),
			NewFooServer("eve"),
			NewFooServer("frank"),
		), nil
	} else if _, ok := st.IsFailover(); ok {
		return nil, errors.New("failover not supported in FooApplication")
	} else if _, ok := st.IsTakeover(); ok {
		return nil, errors.New("takeover not supported in FooApplication")
	}
	panic("unknown start type: " + st.String())
}
