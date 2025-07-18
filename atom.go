package gotp

import (
	"log/slog"
	"unsafe"
)

type Atom string

func (a Atom) String() string {
	return unsafe.String(unsafe.StringData(string(a)), len(a))
}

func (a Atom) Error() string {
	return a.String()
}

func (a Atom) Is(err error) bool {
	if other, ok := err.(Atom); ok {
		return a == other
	}
	return false
}

func (a Atom) LogValue() slog.Value {
	return slog.StringValue(a.String())
}
