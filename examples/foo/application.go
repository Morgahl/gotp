package foo

import (
	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/application"
	"github.com/Morgahl/gotp/server"
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
	srvr := server.New(&NoOp[gotp.Msg, gotp.Msg, gotp.Msg, gotp.Msg, any]{})
	static := supervisor.Static(supervisor.Flags{}, srvr)
	root := supervisor.Static(supervisor.Flags{}, static)
	return root, nil
}
