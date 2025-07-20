package game

import (
	"errors"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/application"
	"github.com/Morgahl/gotp/supervisor"
)

type Game struct{}

func (Game) Name() gotp.Atom {
	return "Game"
}

func (Game) Version() application.Version {
	return application.Ver(0, 1, 0, "alpha")
}

func (f Game) Start(st application.StartType) (supervisor.Supervisable, error) {
	if st.IsNormal() {
		// return treeOfTeams("quartermaster", "teams", [][]gotp.Atom{greekList[:1]}, namesList[:1]), nil
		// return treeOfTeams("quartermaster", "teams", [][]gotp.Atom{greekList[:10]}, namesList[:10]), nil
		// return treeOfTeams("quartermaster", "teams", [][]gotp.Atom{greekList}, namesList), nil
		// return treeOfTeams("quartermaster", "teams", [][]gotp.Atom{greekList, greekList}, namesList), nil
		return treeOfTeams("quartermaster", "teams", [][]gotp.Atom{greekList, greekList, greekList}, namesList), nil
		// return treeOfTeams("quartermaster", "teams", [][]gotp.Atom{greekList, greekList, greekList, greekList[:len(greekList)/4]}, namesList), nil
	} else if _, ok := st.IsFailover(); ok {
		return nil, errors.New("failover not supported in Game")
	} else if _, ok := st.IsTakeover(); ok {
		return nil, errors.New("takeover not supported in Game")
	}
	return nil, errors.New("unknown start type")
}
