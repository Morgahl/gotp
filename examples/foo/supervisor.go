package foo

import (
	"log/slog"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/supervisor"
)

var _ supervisor.Supervisor = &FooSupervisor{}

type FooSupervisor struct {
	name  string
	flags supervisor.Flags
	specs []gotp.Supervisable
}

func NewSupervisor(name string, flags supervisor.Flags, specs ...gotp.Supervisable) gotp.Supervisable {
	sup := &FooSupervisor{
		name:  name,
		flags: flags,
		specs: specs,
	}
	return supervisor.Static(sup)
}

func (f *FooSupervisor) ChildSpec() gotp.ChildSpec {
	return gotp.ChildSpec{
		Name:        f.name,
		Restart:     gotp.PERMANENT,
		Shutdown:    30 * time.Second,
		Type:        gotp.SUPERVISOR,
		Significant: true,
	}
}

func (f *FooSupervisor) Init(opts gotp.Options) (supervisor.Flags, []gotp.Supervisable, error) {
	slog.Debug("FooSupervisor.Init called", "name", f.name, "specs", len(f.specs), "opts", opts)
	return f.flags, f.specs, nil
}
