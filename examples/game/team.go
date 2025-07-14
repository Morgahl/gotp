package game

import (
	"context"
	"log/slog"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/process"
	"github.com/Morgahl/gotp/server"
	"github.com/Morgahl/gotp/supervisor"
)

var _ supervisor.Supervisor[gotp.Options] = &Team{}

type Team struct {
	id     gotp.Atom
	flags  supervisor.Flags
	specs  []server.Supervisable
	server *supervisor.StaticSupervisor
}

func NewTeam(id gotp.Atom, flags supervisor.Flags, specs ...server.Supervisable) server.Supervisable {
	sup := &Team{
		id:    id,
		flags: flags,
		specs: specs,
	}
	sup.server = supervisor.Static(id, sup, nil)
	return sup.server
}

func (f *Team) Context() context.Context {
	return f.server.Context()
}

func (f *Team) ChildSpec() server.ChildSpec {
	return server.ChildSpec{
		ID:          f.id,
		Restart:     server.PERMANENT,
		Shutdown:    0,
		Type:        server.SUPERVISOR,
		Significant: true,
		SpawnOpts:   []process.SpawnOpt{process.Named(f.id)},
	}
}

func (f *Team) Init(opts gotp.Options) (supervisor.Flags, []server.Supervisable, error) {
	slog.InfoContext(f.server.Context(), "Team.Init", "members", len(f.specs), "opts", opts)
	return f.flags, f.specs, nil
}
