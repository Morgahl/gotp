package foo

import (
	"log/slog"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/supervisor"
)

var _ supervisor.Supervisor[gotp.Options] = &FooSupervisor{}

type FooSupervisor struct {
	id    gotp.Atom
	flags supervisor.Flags
	specs []gotp.Supervisable
}

func NewFooSupervisor(id gotp.Atom, flags supervisor.Flags, specs ...gotp.Supervisable) gotp.Supervisable {
	sup := &FooSupervisor{
		id:    id,
		flags: flags,
		specs: specs,
	}
	return supervisor.Static(id, sup, nil)
}

func (f *FooSupervisor) ChildSpec() gotp.ChildSpec {
	return gotp.ChildSpec{
		ID:          f.id,
		Restart:     gotp.PERMANENT,
		Shutdown:    30 * time.Second,
		Type:        gotp.SUPERVISOR,
		Significant: true,
	}
}

func (f *FooSupervisor) Init(opts gotp.Options) (supervisor.Flags, []gotp.Supervisable, error) {
	slog.Info("FooSupervisor.Init", "id", f.id, "specs", len(f.specs), "opts", opts)
	return f.flags, f.specs, nil
}
