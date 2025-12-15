package game

import "github.com/Morgahl/gotp"

var greekCapitalList = []gotp.Atom{
	"Α", // "alpha",
	"Β", // "beta",
	"Γ", // "gamma",
	"Δ", // "delta",
	"Ε", // "epsilon",
	"Ζ", // "zeta",
	"Η", // "eta",
	"Θ", // "theta",
	"Ι", // "iota",
	"Κ", // "kappa",
	"Λ", // "lambda",
	"Μ", // "mu",
	"Ν", // "nu",
	"Ξ", // "xi",
	"Ο", // "omicron",
	"Π", // "pi",
	"Ρ", // "rho",
	"Σ", // "sigma",
	"Τ", // "tau",
	"Υ", // "upsilon",
	"Φ", // "phi",
	"Χ", // "chi",
	"Ψ", // "psi",
	"Ω", // "omega",
}

var greekLowerList = []gotp.Atom{
	"α", // "alpha",
	"β", // "beta",
	"γ", // "gamma",
	"δ", // "delta",
	"ε", // "epsilon",
	"ζ", // "zeta",
	"η", // "eta",
	"θ", // "theta",
	"ι", // "iota",
	"κ", // "kappa",
	"λ", // "lambda",
	"μ", // "mu",
	"ν", // "nu",
	"ξ", // "xi",
	"ο", // "omicron",
	"π", // "pi",
	"ρ", // "rho",
	"σ", // "sigma",
	"τ", // "tau",
	"υ", // "upsilon",
	"φ", // "phi",
	"χ", // "chi",
	"ψ", // "psi",
	"ω", // "omega",
}

var greekGreekList = permuteLists(greekCapitalList, greekCapitalList)
var greekGreekGreekList = permuteLists(greekGreekList, greekCapitalList)

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

var greekNamesList = permuteLists(greekCapitalList, namesList)

func permuteLists(left, right []gotp.Atom) []gotp.Atom {
	list := make([]gotp.Atom, 0, len(left)*len(right))
	for _, l := range left {
		for _, r := range right {
			list = append(list, gotp.Atom(l+"_"+r))
		}
	}
	return list
}
