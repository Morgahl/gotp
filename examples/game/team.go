package game

import (
	"log/slog"

	gotp "github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/server"
	"github.com/Morgahl/gotp/supervisor"
)

var _ supervisor.Supervisor[gotp.Options] = &Team{}

type Team struct {
	id    gotp.Atom
	flags supervisor.Flags
	specs []server.Supervisable
}

func NewTeam(id gotp.Atom, flags supervisor.Flags, specs ...server.Supervisable) server.Supervisable {
	sup := &Team{
		id:    id,
		flags: flags,
		specs: specs,
	}
	return supervisor.Static(id, sup, nil)
}

func (f *Team) ChildSpec() server.ChildSpec {
	return server.ChildSpec{
		ID:          f.id,
		Restart:     server.TRANSIENT,
		Shutdown:    gotp.DEFAULT_SHUTDOWN,
		Type:        server.SUPERVISOR,
		Significant: true,
	}
}

func (f *Team) Init(opts gotp.Options) (supervisor.Flags, []server.Supervisable, error) {
	slog.Info("Team.Init", "id", f.id, "members", len(f.specs), "opts", opts)
	return f.flags, f.specs, nil
}
