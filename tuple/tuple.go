package tuple

import "github.com/Morgahl/gotp"

var _ gotp.Term = T0{}
var _ gotp.Term = T1[gotp.Term]{}
var _ gotp.Term = T2[gotp.Term, gotp.Term]{}
var _ gotp.Term = T3[gotp.Term, gotp.Term, gotp.Term]{}
var _ gotp.Term = T4[gotp.Term, gotp.Term, gotp.Term, gotp.Term]{}
var _ gotp.Term = T5[gotp.Term, gotp.Term, gotp.Term, gotp.Term, gotp.Term]{}
var _ gotp.Term = T6[gotp.Term, gotp.Term, gotp.Term, gotp.Term, gotp.Term, gotp.Term]{}
var _ gotp.Term = T7[gotp.Term, gotp.Term, gotp.Term, gotp.Term, gotp.Term, gotp.Term, gotp.Term]{}
var _ gotp.Term = T8[gotp.Term, gotp.Term, gotp.Term, gotp.Term, gotp.Term, gotp.Term, gotp.Term, gotp.Term]{}
var _ gotp.Term = T9[gotp.Term, gotp.Term, gotp.Term, gotp.Term, gotp.Term, gotp.Term, gotp.Term, gotp.Term, gotp.Term]{}
var _ gotp.Term = T10[gotp.Term, gotp.Term, gotp.Term, gotp.Term, gotp.Term, gotp.Term, gotp.Term, gotp.Term, gotp.Term, gotp.Term]{}
var _ gotp.Term = T11[gotp.Term, gotp.Term, gotp.Term, gotp.Term, gotp.Term, gotp.Term, gotp.Term, gotp.Term, gotp.Term, gotp.Term, gotp.Term]{}

type T0 struct{}

type T1[T0 gotp.Term] struct {
	T0 T0
}

type T2[T0, T1 gotp.Term] struct {
	T0 T0
	T1 T1
}

type T3[T0, T1, T2 gotp.Term] struct {
	T0 T0
	T1 T1
	T2 T2
}

type T4[T0, T1, T2, T3 gotp.Term] struct {
	T0 T0
	T1 T1
	T2 T2
	T3 T3
}

type T5[T0, T1, T2, T3, T4 gotp.Term] struct {
	T0 T0
	T1 T1
	T2 T2
	T3 T3
	T4 T4
}

type T6[T0, T1, T2, T3, T4, T5 gotp.Term] struct {
	T0 T0
	T1 T1
	T2 T2
	T3 T3
	T4 T4
	T5 T5
}

type T7[T0, T1, T2, T3, T4, T5, T6 gotp.Term] struct {
	T0 T0
	T1 T1
	T2 T2
	T3 T3
	T4 T4
	T5 T5
	T6 T6
}

type T8[T0, T1, T2, T3, T4, T5, T6, T7 gotp.Term] struct {
	T0 T0
	T1 T1
	T2 T2
	T3 T3
	T4 T4
	T5 T5
	T6 T6
	T7 T7
}

type T9[T0, T1, T2, T3, T4, T5, T6, T7, T8 gotp.Term] struct {
	T0 T0
	T1 T1
	T2 T2
	T3 T3
	T4 T4
	T5 T5
	T6 T6
	T7 T7
	T8 T8
}

type T10[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9 gotp.Term] struct {
	T0 T0
	T1 T1
	T2 T2
	T3 T3
	T4 T4
	T5 T5
	T6 T6
	T7 T7
	T8 T8
	T9 T9
}

type T11[T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10 gotp.Term] struct {
	T0  T0
	T1  T1
	T2  T2
	T3  T3
	T4  T4
	T5  T5
	T6  T6
	T7  T7
	T8  T8
	T9  T9
	T10 T10
}
