package game

import "github.com/Morgahl/gotp"

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

var greekGreekList = permuteLists(greekList, greekList)
var greekGreekGreekList = permuteLists(greekGreekList, greekList)

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

func permuteLists(left, right []gotp.Atom) []gotp.Atom {
	list := make([]gotp.Atom, 0, len(left)*len(right))
	for _, l := range left {
		for _, r := range right {
			list = append(list, gotp.Atom(l+"_"+r))
		}
	}
	return list
}
