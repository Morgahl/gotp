package gotp

import "unsafe"

type Atom string

func (a Atom) String() string {
	return unsafe.String(unsafe.StringData(string(a)), len(a))
}
