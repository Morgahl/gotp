package foo

import (
	"log/slog"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/supervisor"
)

var _ supervisor.Supervisor[gotp.Options] = &Team{}

type Team struct {
	id    gotp.Atom
	flags supervisor.Flags
	specs []gotp.Supervisable
}

func NewTeam(id gotp.Atom, flags supervisor.Flags, specs ...gotp.Supervisable) gotp.Supervisable {
	sup := &Team{
		id:    id,
		flags: flags,
		specs: specs,
	}
	return supervisor.Static(id, sup, nil)
}

func (f *Team) ChildSpec() gotp.ChildSpec {
	return gotp.ChildSpec{
		ID:          f.id,
		Restart:     gotp.TRANSIENT,
		Shutdown:    30 * time.Second,
		Type:        gotp.SUPERVISOR,
		Significant: true,
	}
}

func (f *Team) Init(opts gotp.Options) (supervisor.Flags, []gotp.Supervisable, error) {
	slog.Info("Team.Init", "id", f.id, "members", len(f.specs), "opts", opts)
	return f.flags, f.specs, nil
}
