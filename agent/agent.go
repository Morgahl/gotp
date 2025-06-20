package agent

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/server"
)

var _ server.Serverable[any, gotp.Msg, int, gotp.Msg, gotp.Msg] = &Agent[int]{}

type InitFn[T any] func() *T

type GetFn[T any] func(T) T

type GetAndUpdateFn[T any] func(*T) T

type UpdateFn[T any] func(*T)

type Agent[T any] struct {
	initFn InitFn[T]
	state  *T

	server.OptionalCallbacks[any, gotp.Msg]
	server *server.Server[any, gotp.Msg, T, gotp.Msg, gotp.Msg]
}

func New[T any](fn InitFn[T]) *Agent[T] {
	a := &Agent[T]{initFn: fn}
	a.server = server.New(a, nil)
	return a
}

func (a *Agent[T]) Start(opts ...gotp.SpawnOpt) (gotp.Started, error) {
	return a.server.Start(opts...)
}

func (a *Agent[T]) StartLink(link gotp.PID, opts ...gotp.SpawnOpt) (gotp.Supervised, error) {
	return a.server.StartLink(link, opts...)
}

func (a *Agent[T]) ChildSpec() gotp.ChildSpec {
	return gotp.ChildSpec{
		Restart:     gotp.PERMANENT,
		Shutdown:    30 * time.Second,
		Type:        gotp.WORKER,
		Significant: true,
	}
}

func (a *Agent[T]) Init(any) (server.Continue[gotp.Msg], error) {
	if a.initFn == nil {
		return server.NoCont[gotp.Msg](), fmt.Errorf("init function cannot be nil")
	}
	a.state = a.initFn()
	return server.NoCont[gotp.Msg](), nil
}

func (a *Agent[T]) HandleCall(msg gotp.Msg, from gotp.PID) (server.Response[T], server.Continue[gotp.Msg], error) {
	slog.Debug("HandleCall", slog.String("msg", fmt.Sprintf("%+v", msg)), slog.Any("from", from))
	switch msg := msg.(type) {
	case GetFn[T]:
		return server.Reply(msg(*a.state)), server.NoCont[gotp.Msg](), nil
	case GetAndUpdateFn[T]:
		return server.Reply(msg(a.state)), server.NoCont[gotp.Msg](), nil
	case UpdateFn[T]:
		msg(a.state)
		return server.NoReply[T](), server.NoCont[gotp.Msg](), nil
	default:
		cont, err := a.HandleInfo(msg)
		return server.NoReply[T](), cont, err
	}
}

func (a *Agent[T]) HandleCast(msg gotp.Msg) (server.Continue[gotp.Msg], error) {
	slog.Debug("HandleCast", slog.String("msg", fmt.Sprintf("%+v", msg)))
	switch msg := msg.(type) {
	case UpdateFn[T]:
		msg(a.state)
		return server.NoCont[gotp.Msg](), nil
	default:
		return a.HandleInfo(msg)
	}
}
