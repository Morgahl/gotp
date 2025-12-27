package tuple

import (
	"encoding/gob"

	"github.com/Morgahl/gotp/term"
)

func init() {
	gob.Register(T0{})
	gob.Register(T1[term.Term]{})
	gob.Register(T2[term.Term, term.Term]{})
	gob.Register(T3[term.Term, term.Term, term.Term]{})
	gob.Register(T4[term.Term, term.Term, term.Term, term.Term]{})
	gob.Register(T5[term.Term, term.Term, term.Term, term.Term, term.Term]{})
	gob.Register(T6[term.Term, term.Term, term.Term, term.Term, term.Term, term.Term]{})
	gob.Register(T7[term.Term, term.Term, term.Term, term.Term, term.Term, term.Term, term.Term]{})
	gob.Register(T8[term.Term, term.Term, term.Term, term.Term, term.Term, term.Term, term.Term, term.Term]{})
	gob.Register(T9[term.Term, term.Term, term.Term, term.Term, term.Term, term.Term, term.Term, term.Term, term.Term]{})
	gob.Register(T10[term.Term, term.Term, term.Term, term.Term, term.Term, term.Term, term.Term, term.Term, term.Term, term.Term]{})
	gob.Register(T11[term.Term, term.Term, term.Term, term.Term, term.Term, term.Term, term.Term, term.Term, term.Term, term.Term, term.Term]{})
}

var _ term.Term = &T0{}
var _ term.Term = &T1[term.Term]{}
var _ term.Term = &T2[term.Term, term.Term]{}
var _ term.Term = &T3[term.Term, term.Term, term.Term]{}
var _ term.Term = &T4[term.Term, term.Term, term.Term, term.Term]{}
var _ term.Term = &T5[term.Term, term.Term, term.Term, term.Term, term.Term]{}
var _ term.Term = &T6[term.Term, term.Term, term.Term, term.Term, term.Term, term.Term]{}
var _ term.Term = &T7[term.Term, term.Term, term.Term, term.Term, term.Term, term.Term, term.Term]{}
var _ term.Term = &T8[term.Term, term.Term, term.Term, term.Term, term.Term, term.Term, term.Term, term.Term]{}
var _ term.Term = &T9[term.Term, term.Term, term.Term, term.Term, term.Term, term.Term, term.Term, term.Term, term.Term]{}
var _ term.Term = &T10[term.Term, term.Term, term.Term, term.Term, term.Term, term.Term, term.Term, term.Term, term.Term, term.Term]{}
var _ term.Term = &T11[term.Term, term.Term, term.Term, term.Term, term.Term, term.Term, term.Term, term.Term, term.Term, term.Term, term.Term]{}

type T0 struct{}

type T1[E0 term.Term] struct{ E0 E0 }

type T2[E0, E1 term.Term] struct {
	E0 E0
	E1 E1
}

type T3[E0, E1, E2 term.Term] struct {
	E0 E0
	E1 E1
	E2 E2
}

type T4[E0, E1, E2, E3 term.Term] struct {
	E0 E0
	E1 E1
	E2 E2
	E3 E3
}

type T5[E0, E1, E2, E3, E4 term.Term] struct {
	E0 E0
	E1 E1
	E2 E2
	E3 E3
	E4 E4
}

type T6[E0, E1, E2, E3, E4, E5 term.Term] struct {
	E0 E0
	E1 E1
	E2 E2
	E3 E3
	E4 E4
	E5 E5
}

type T7[E0, E1, E2, E3, E4, E5, E6 term.Term] struct {
	E0 E0
	E1 E1
	E2 E2
	E3 E3
	E4 E4
	E5 E5
	E6 E6
}

type T8[E0, E1, E2, E3, E4, E5, E6, E7 term.Term] struct {
	E0 E0
	E1 E1
	E2 E2
	E3 E3
	E4 E4
	E5 E5
	E6 E6
	E7 E7
}

type T9[E0, E1, E2, E3, E4, E5, E6, E7, E8 term.Term] struct {
	E0 E0
	E1 E1
	E2 E2
	E3 E3
	E4 E4
	E5 E5
	E6 E6
	E7 E7
	E8 E8
}

type T10[E0, E1, E2, E3, E4, E5, E6, E7, E8, E9 term.Term] struct {
	E0 E0
	E1 E1
	E2 E2
	E3 E3
	E4 E4
	E5 E5
	E6 E6
	E7 E7
	E8 E8
	E9 E9
}

type T11[E0, E1, E2, E3, E4, E5, E6, E7, E8, E9, E10 term.Term] struct {
	E0  E0
	E1  E1
	E2  E2
	E3  E3
	E4  E4
	E5  E5
	E6  E6
	E7  E7
	E8  E8
	E9  E9
	E10 E10
}
