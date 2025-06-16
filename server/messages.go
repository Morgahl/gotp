package server

import (
	"github.com/Morgahl/gotp"
)

type respEnum uint8

const (
	NO_REPLY respEnum = iota
	REPLY
)

type Response[R any] struct {
	atom respEnum
	resp R
}

func NoReply[R any]() Response[R] {
	return Response[R]{atom: NO_REPLY}
}

func Reply[R any](resp R) Response[R] {
	return Response[R]{atom: REPLY, resp: resp}
}

func (r Response[R]) IsReply() (R, bool) {
	if r.atom == REPLY {
		return r.resp, true
	}
	var zero R
	return zero, false
}

func (r Response[R]) IsNoReply() bool {
	return r.atom == NO_REPLY
}

type contAtom uint8

const (
	NO_CONTINUE contAtom = iota
	CONTINUE
)

type Continue[C any] struct {
	atom contAtom
	arg  C
}

func NoCont[C any]() Continue[C] {
	return Continue[C]{atom: NO_CONTINUE}
}

func Cont[C any](contArg C) Continue[C] {
	return Continue[C]{atom: CONTINUE, arg: contArg}
}

type call[M gotp.Msg, R gotp.Msg] struct {
	req  M
	from gotp.PID
	resp chan R
}

// TODO: Convert this from Call creator to the Call function that handles calling the server if it is a server or by pid
// TODO: to send the call message and handle the respnose waitng AS WELL AS THE close of the chanle without response
// TODO: indicating NO_REPLY.
func Call[M gotp.Msg, R gotp.Msg](req M, from gotp.PID) call[M, R] {
	return call[M, R]{req: req, from: from, resp: make(chan R, 1)}
}

type cast[M gotp.Msg] struct {
	cast M
}

// TODO: Convert this from Cast creator to the Cast function that handles calling the server if it is a server or by pid
// TODO: to send the cast message and do not wait for the response
func Cast[M gotp.Msg](req M) cast[M] {
	return cast[M]{cast: req}
}
