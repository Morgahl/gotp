package game

import (
	"errors"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/application"
	"github.com/Morgahl/gotp/supervisor"
)

func init() {
	// TODO: let's make it so the GRTS runs applications are "Registered" during an init like this
}

type Game struct{}

func (Game) Name() gotp.Atom {
	return "Game"
}

func (Game) Version() application.Version {
	return application.Ver(0, 1, 0, "alpha")
}

func (f Game) Start(st application.StartType) (supervisor.Supervisable, error) {
	if st.IsNormal() {
		// return treeOfTeams("quartermaster", "teams", [][]gotp.Atom{greekCapitalList[:1]}, namesList[:1]), nil // 5 processes
		// return treeOfTeams("quartermaster", "teams", [][]gotp.Atom{greekCapitalList[:10]}, namesList[:10]), nil // 113 processes
		// return treeOfTeams("quartermaster", "teams", [][]gotp.Atom{greekCapitalList}, namesList), nil // 651 processes
		// return treeOfTeams("quartermaster", "teams", [][]gotp.Atom{greekCapitalList, greekCapitalList[:16]}, namesList), nil // 10,395 processes
		// return treeOfTeams("quartermaster", "teams", [][]gotp.Atom{greekCapitalList, greekCapitalList, greekCapitalList[:4]}, namesList), nil // 62,811 processes
		// return treeOfTeams("quartermaster", "teams", [][]gotp.Atom{greekCapitalList, greekCapitalList, greekCapitalList[:7]}, namesList), nil // 109,467 processes
		// return treeOfTeams("quartermaster", "teams", [][]gotp.Atom{greekCapitalList, greekCapitalList, greekCapitalList}, namesList), nil // 373,851 processes
		// return treeOfTeams("quartermaster", "teams", [][]gotp.Atom{greekCapitalList, greekCapitalList, greekCapitalList, greekCapitalList[:2]}, namesList), nil // 760,923 processes
		// return treeOfTeams("quartermaster", "teams", [][]gotp.Atom{greekCapitalList, greekCapitalList, greekCapitalList, greekCapitalList[:3]}, namesList), nil // 1,134,171 processes
		// return treeOfTeams("quartermaster", "teams", [][]gotp.Atom{greekCapitalList, greekCapitalList, greekCapitalList, greekCapitalList[:4]}, namesList), nil // 1,507,419 processes
		// return treeOfTeams("quartermaster", "teams", [][]gotp.Atom{greekCapitalList, greekCapitalList, greekCapitalList, greekCapitalList[:5]}, namesList), nil // 1,880,667 processes
		// return treeOfTeams("quartermaster", "teams", [][]gotp.Atom{greekCapitalList, greekCapitalList, greekCapitalList, greekCapitalList[:6]}, namesList), nil // 2,253,915 processes
		return treeOfTeams("quartermaster", "teams", [][]gotp.Atom{greekCapitalList, greekCapitalList, greekCapitalList, greekCapitalList[:7]}, namesList), nil // 2,627,163 processes
		// return treeOfTeams("quartermaster", "teams", [][]gotp.Atom{greekCapitalList, greekCapitalList, greekCapitalList, greekCapitalList[:8]}, namesList), nil
	} else if _, ok := st.IsFailover(); ok {
		return nil, errors.New("failover not supported in Game")
	} else if _, ok := st.IsTakeover(); ok {
		return nil, errors.New("takeover not supported in Game")
	}
	return nil, errors.New("unknown start type")
}
