package game

import (
	"errors"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/application"
	"github.com/Morgahl/gotp/debug"
	"github.com/Morgahl/gotp/supervisor"
)

type Game struct{}

func (Game) Name() gotp.Atom {
	return "Game"
}

func (Game) Version() application.Version {
	return application.Ver(0, 1, 0, "alpha")
}

func (f Game) Start(st application.StartType) (gotp.Supervisable, error) {
	if st.IsNormal() {
		return NewTeam("teams", supervisor.Flags{},
			NewTeam("alpha", supervisor.Flags{}, f.assignCrew("alpha")...),
			NewTeam("beta", supervisor.Flags{}, f.assignCrew("beta")...),
			NewTeam("gamma", supervisor.Flags{}, f.assignCrew("gamma")...),
			NewTeam("delta", supervisor.Flags{}, f.assignCrew("delta")...),
			NewTeam("epsilon", supervisor.Flags{}, f.assignCrew("epsilon")...),
			NewTeam("zeta", supervisor.Flags{}, f.assignCrew("zeta")...),
			NewTeam("eta", supervisor.Flags{}, f.assignCrew("eta")...),
			NewTeam("theta", supervisor.Flags{}, f.assignCrew("theta")...),
			NewTeam("iota", supervisor.Flags{}, f.assignCrew("iota")...),
			NewTeam("kappa", supervisor.Flags{}, f.assignCrew("kappa")...),
			NewTeam("lambda", supervisor.Flags{}, f.assignCrew("lambda")...),
			NewTeam("mu", supervisor.Flags{}, f.assignCrew("mu")...),
			NewTeam("nu", supervisor.Flags{}, f.assignCrew("nu")...),
			NewTeam("xi", supervisor.Flags{}, f.assignCrew("xi")...),
			NewTeam("omicron", supervisor.Flags{}, f.assignCrew("omicron")...),
			NewTeam("pi", supervisor.Flags{}, f.assignCrew("pi")...),
			NewTeam("rho", supervisor.Flags{}, f.assignCrew("rho")...),
			NewTeam("sigma", supervisor.Flags{}, f.assignCrew("sigma")...),
			NewTeam("tau", supervisor.Flags{}, f.assignCrew("tau")...),
			NewTeam("upsilon", supervisor.Flags{}, f.assignCrew("upsilon")...),
			NewTeam("phi", supervisor.Flags{}, f.assignCrew("phi")...),
			NewTeam("chi", supervisor.Flags{}, f.assignCrew("chi")...),
			NewTeam("psi", supervisor.Flags{}, f.assignCrew("psi")...),
			NewTeam("omega", supervisor.Flags{}, f.assignCrew("omega")...),
		), nil
	} else if _, ok := st.IsFailover(); ok {
		return nil, errors.New("failover not supported in Game")
	} else if _, ok := st.IsTakeover(); ok {
		return nil, errors.New("takeover not supported in Game")
	}
	panic(debug.ThrowF("unknown start type: %s", st))
}

func (Game) assignCrew(n gotp.Atom) []gotp.Supervisable {
	return []gotp.Supervisable{
		NewCrew(n + "_alice"),
		NewCrew(n + "_bob"),
		NewCrew(n + "_charlie"),
		NewCrew(n + "_dave"),
		NewCrew(n + "_eve"),
		NewCrew(n + "_frank"),
		NewCrew(n + "_grace"),
		NewCrew(n + "_heidi"),
		NewCrew(n + "_ivan"),
		NewCrew(n + "_judy"),
		NewCrew(n + "_ken"),
		NewCrew(n + "_larry"),
		NewCrew(n + "_mallory"),
		NewCrew(n + "_nina"),
		NewCrew(n + "_oscar"),
		NewCrew(n + "_peter"),
		NewCrew(n + "_quinn"),
		NewCrew(n + "_rachel"),
		NewCrew(n + "_steve"),
		NewCrew(n + "_trudy"),
		NewCrew(n + "_victor"),
		NewCrew(n + "_wendy"),
		NewCrew(n + "_xander"),
		NewCrew(n + "_yara"),
		NewCrew(n + "_zara"),
	}
}
