package server

import "github.com/Morgahl/gotp"

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

type callMsg[M gotp.Msg, R gotp.Msg] struct {
	req  M
	from gotp.PID
	resp chan R
}

func newCallMsg[M gotp.Msg, R gotp.Msg](req M, from gotp.PID) callMsg[M, R] {
	return callMsg[M, R]{req: req, from: from, resp: make(chan R, 1)}
}

type castMsg[M gotp.Msg] struct {
	cast M
}

func newCastMsg[M gotp.Msg](req M) castMsg[M] {
	return castMsg[M]{cast: req}
}

type infoMsg[M gotp.Msg] struct {
	info M
}

func newInfoMsg[M gotp.Msg](info M) infoMsg[M] {
	return infoMsg[M]{info: info}
}
