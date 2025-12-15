package gotp

import (
	"encoding/gob"
	"log/slog"
)

func init() {
	gob.Register(Atom(""))
}

type Atom string

func (a Atom) String() string {
	// return unsafe.String(unsafe.StringData(string(a)), len(a))
	return string(a)
}

func (a Atom) Error() string {
	// return a.String()
	return string(a)
}

func (a Atom) Is(err error) bool {
	other, ok := err.(Atom)
	return ok && a == other
}

func (a Atom) LogValue() slog.Value {
	// return slog.StringValue(a.String())
	return slog.StringValue(string(a))
}
