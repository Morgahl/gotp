package node

import (
	"errors"

	"github.com/Morgahl/gotp"
)

type Local struct {
	Node
}

func newLocal(sname gotp.Atom, cookie gotp.Atom) *Local {
	return &Local{
		Node: Node{
			Name:   sname,
			Cookie: Cookie(cookie),
		},
	}
}

type Args struct {
	A, B int
}

type Quotient struct {
	Quo, Rem int
}

func (t *Local) Multiply(args *Args, reply *int) error {
	*reply = args.A * args.B
	return nil
}

func (t *Local) Divide(args *Args, quo *Quotient) error {
	if args.B == 0 {
		return errors.New("divide by zero")
	}
	quo.Quo = args.A / args.B
	quo.Rem = args.A % args.B
	return nil
}
