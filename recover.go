package gotp

import "fmt"

func Recover(err *error) {
	if r := recover(); r != nil {
		*err = newRecovered(*err, r)
	}
}

type Recovered struct {
	err error
	r   error
}

func newRecovered(err error, r any) Recovered {
	if r, ok := r.(error); ok {
		return Recovered{err: err, r: fmt.Errorf("panic: %w", r)}
	}
	return Recovered{err: err, r: fmt.Errorf("panic: %v", r)}
}

func (r Recovered) Error() string {
	if r.err == nil {
		return fmt.Sprintf("Recovered{%v}", r.r)
	}
	return fmt.Sprintf("Recovered{%v, %v}", r.err, r.r)
}
