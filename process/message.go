package process

import "fmt"

type Message interface{}

type Request[M Message] struct {
	From    PID
	Ref     *Ref
	Message M
}

type Reply[M Message] struct {
	From    PID
	Ref     *Ref
	Message M
}

type Exit struct {
	Sender PID
	Reason fmt.Stringer
}

type Down struct {
	From   PID
	Ref    *Ref
	Reason error
}

type Reason struct {
	error
}

func (r Reason) String() string {
	return r.Error()
}
