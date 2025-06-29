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
	Sender *Ref
	Reason fmt.Stringer
}

type Down struct {
	From   PID
	Ref    *Ref
	Reason error
}
