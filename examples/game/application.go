package game

import (
	"errors"

	gotp "github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/application"
	"github.com/Morgahl/gotp/server"
	"github.com/Morgahl/gotp/supervisor"
)

var greekList = []gotp.Atom{
	"alpha",
	"beta",
	"gamma",
	"delta",
	"epsilon",
	"zeta",
	"eta",
	"theta",
	"iota",
	"kappa",
	"lambda",
	"mu",
	"nu",
	"xi",
	"omicron",
	"pi",
	"rho",
	"sigma",
	"tau",
	"upsilon",
	"phi",
	"chi",
	"psi",
	"omega",
}

var namesList = []gotp.Atom{
	"alice",
	"bob",
	"charlie",
	"dave",
	"eve",
	"frank",
	"grace",
	"heidi",
	"ivan",
	"judy",
	"ken",
	"larry",
	"mallory",
	"nina",
	"oscar",
	"peter",
	"quinn",
	"rachel",
	"steve",
	"trudy",
	"ursula",
	"victor",
	"wendy",
	"xander",
	"yara",
	"zara",
}

var greekNamesList = permuteLists(greekList, namesList)

var namesGreekList = permuteLists(namesList, greekList)

type Game struct{}

func (Game) Name() gotp.Atom {
	return "Game"
}

func (Game) Version() application.Version {
	return application.Ver(0, 1, 0, "alpha")
}

func (f Game) Start(st application.StartType) (server.Supervisable, error) {
	if st.IsNormal() {
		return NewTeam("teams", supervisor.Flags{}, buildTeams(greekList[:1], namesList[:2])...), nil
		// return NewTeam("teams", supervisor.Flags{}, buildTeams(greekList, namesList)...), nil
		// return NewTeam("teams", supervisor.Flags{}, buildTeams(greekNamesList, greekNamesList)...), nil
	} else if _, ok := st.IsFailover(); ok {
		return nil, errors.New("failover not supported in Game")
	} else if _, ok := st.IsTakeover(); ok {
		return nil, errors.New("takeover not supported in Game")
	}
	return nil, errors.New("unknown start type")
}

func permuteLists(left, right []gotp.Atom) []gotp.Atom {
	list := make([]gotp.Atom, 0, len(left)*len(right))
	for _, l := range left {
		for _, r := range right {
			list = append(list, gotp.Atom(l+"_"+r))
		}
	}
	return list
}

func buildTeams(teams, crew []gotp.Atom) []server.Supervisable {
	var supervisors []server.Supervisable
	for _, t := range teams {
		supervisors = append(supervisors, NewTeam(t, supervisor.Flags{}, buildCrew(t, crew)...))
	}
	return supervisors
}

func buildCrew(team gotp.Atom, crewList []gotp.Atom) []server.Supervisable {
	var crew []server.Supervisable
	for _, c := range crewList {
		crew = append(crew, NewCrew(team+"_"+c))
	}
	return crew
}
