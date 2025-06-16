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
		return NewFooSupervisor("teams", supervisor.Flags{},
			NewFooSupervisor("alpha", supervisor.Flags{},
				NewFooServer("alice"),
				NewFooServer("bob"),
				NewFooServer("charlie")),
			NewFooSupervisor("beta", supervisor.Flags{},
				NewFooServer("dave"),
				NewFooServer("eve"),
				NewFooServer("frank")),
			NewFooSupervisor("gamma", supervisor.Flags{},
				NewFooServer("grace"),
				NewFooServer("heidi"),
				NewFooServer("ivan")),
			NewFooSupervisor("delta", supervisor.Flags{},
				NewFooServer("judy"),
				NewFooServer("ken"),
				NewFooServer("larry")),
			NewFooSupervisor("epsilon", supervisor.Flags{},
				NewFooServer("mallory"),
				NewFooServer("nina"),
				NewFooServer("oscar")),
			NewFooSupervisor("zeta", supervisor.Flags{},
				NewFooServer("peter"),
				NewFooServer("quinn"),
				NewFooServer("rachel")),
			NewFooSupervisor("eta", supervisor.Flags{},
				NewFooServer("steve"),
				NewFooServer("trudy"),
				NewFooServer("victor")),
			NewFooSupervisor("theta", supervisor.Flags{},
				NewFooServer("wendy"),
				NewFooServer("xander"),
				NewFooServer("yara")),
			NewFooSupervisor("omega", supervisor.Flags{},
				NewFooServer("zara"),
			),
		), nil
	} else if _, ok := st.IsFailover(); ok {
		return nil, errors.New("failover not supported in FooApplication")
	} else if _, ok := st.IsTakeover(); ok {
		return nil, errors.New("takeover not supported in FooApplication")
	}
	panic("unknown start type: " + st.String())
}
