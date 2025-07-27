package process

import "github.com/Morgahl/gotp"

type From interface {
	PID | Ref
}

type RequestMsg[M gotp.Term] struct {
	From    PID
	Ref     Ref
	Message M
}

func RequestFrom[F From, M gotp.Term](from F, msg M) RequestMsg[M] {
	var req RequestMsg[M]
	switch f := any(from).(type) {
	case PID:
		req = RequestMsg[M]{From: f, Message: msg}
	case Ref:
		req = RequestMsg[M]{From: f.pid, Ref: f, Message: msg}
	}
	return req
}

type ReplyMsg[M gotp.Term] struct {
	From    PID
	Ref     Ref
	Message M
}

func ReplyTo[F From, M gotp.Term](request RequestMsg[gotp.Term], from F, msg M) ReplyMsg[M] {
	var rep ReplyMsg[M]
	switch f := any(from).(type) {
	case PID:
		rep = ReplyMsg[M]{From: f, Ref: request.Ref, Message: msg}
	case Ref:
		rep = ReplyMsg[M]{From: f.pid, Ref: f, Message: msg}
	}
	return rep
}

type ExitMsg struct {
	PID    PID
	Reason error
}

type DownMsg struct {
	From   PID
	Ref    Ref
	Reason error
}
