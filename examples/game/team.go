package game

import (
	"log/slog"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/process"
	"github.com/Morgahl/gotp/supervisor"
	"github.com/Morgahl/gotp/supervisor/static"
)

var _ supervisor.Supervisor[gotp.Options] = &Team{}

// for each list of team names we create another tier of teams with the names concatenated with an underscore the last
// set of teams will have crew as their children
func treeOfTeams(workAgent, team gotp.Atom, teams [][]gotp.Atom, crew []gotp.Atom) supervisor.Supervisable {
	var supervisors []supervisor.Supervisable
	if len(teams) == 0 {
		return NewTeam(team, supervisor.Options{}, buildCrew(workAgent, team, crew)...)
	}

	for _, t := range teams[0] {
		if team == "" {
			supervisors = append(supervisors, treeOfTeams(workAgent, t, teams[1:], crew))
		} else {
			supervisors = append(supervisors, treeOfTeams(workAgent, team+"_"+t, teams[1:], crew))
		}
	}
	if team == "" {
		count := len(teams[0])
		for _, t := range teams[1:] {
			count *= len(t)
		}
		// prepend the work agent to the list of supervisors
		supervisors = append([]supervisor.Supervisable{NewWorkAgent(workAgent, 2*uint64(count*len(crew)))}, supervisors...)
	}
	return NewTeam(team, supervisor.Options{}, supervisors...)
}

func buildTeams(workAgent gotp.Atom, teams, crew []gotp.Atom) []supervisor.Supervisable {
	var supervisors []supervisor.Supervisable
	supervisors = append(supervisors, NewWorkAgent(workAgent, 2*uint64(len(teams)*len(crew))))
	for _, t := range teams {
		supervisors = append(supervisors, NewTeam(t, supervisor.Options{}, buildCrew(workAgent, t, crew)...))
	}
	return supervisors
}

type Team struct {
	id    gotp.Atom
	flags supervisor.Options
	specs []supervisor.Supervisable
}

func NewTeam(id gotp.Atom, flags supervisor.Options, specs ...supervisor.Supervisable) *Team {
	return &Team{
		id:    id,
		flags: flags,
		specs: specs,
	}
}

func (f *Team) ChildSpec() supervisor.ChildSpec {
	return supervisor.ChildSpec{
		ID:          f.id,
		Restart:     supervisor.PERMANENT,
		Shutdown:    0,
		Type:        supervisor.SUPERVISOR,
		Significant: true,
		Start: func(opts ...process.SpawnOpt) (process.Ref, error) {
			return static.Start(f, nil, append([]process.SpawnOpt{process.Named(f.id)}, opts...)...)
		},
	}
}

func (f *Team) Init(pctx process.Context, opts gotp.Options) (supervisor.Options, []supervisor.Supervisable, error) {
	slog.InfoContext(pctx.Context(), "Team.Init", "members", len(f.specs), "opts", opts)
	return f.flags, f.specs, nil
}
