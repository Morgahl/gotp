package gen_server

import (
	"fmt"
	"log/slog"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/process"
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
	STOP
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

func Stop[C any](reason C) Continue[C] {
	return Continue[C]{atom: STOP, arg: reason}
}

type call[T gotp.Term, R gotp.Term] struct {
	from process.PID
	req  T
	resp chan R
}

func CallMsg[T gotp.Term, R gotp.Term](from process.PID, req T) call[T, R] {
	return call[T, R]{from: from, req: req, resp: make(chan R, 1)}
}

func (c call[T, R]) String() string {
	return fmt.Sprintf("call{from: %s, req: %T}", c.from, c.req)
}

func (c call[T, R]) LogValue() slog.Value {
	return slog.StringValue(c.String())
}

type cast[T gotp.Term] struct {
	req T
}

func CastMsg[T gotp.Term](req T) cast[T] {
	return cast[T]{req: req}
}

func (c cast[T]) String() string {
	return fmt.Sprintf("cast{req: %T}", c.req)
}

func (c cast[T]) LogValue() slog.Value {
	return slog.StringValue(c.String())
}

type stop struct {
	reason error
}

func StopMsg(reason error) stop {
	return stop{reason: reason}
}

func (s stop) String() string {
	return fmt.Sprintf("stop{reason: %s}", s.reason)
}

func (s stop) LogValue() slog.Value {
	return slog.StringValue(s.String())
}
