package agent

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/gen_server"
)

var _ gen_server.Serverable[any, gotp.Msg, int, gotp.Msg, gotp.Msg] = &Agent[int]{}

type InitFn[T any] func() *T

type GetFn[T any] func(T) T

type GetAndUpdateFn[T any] func(*T) T

type UpdateFn[T any] func(*T)

type Agent[T any] struct {
	initFn InitFn[T]
	state  *T

	gen_server.OptionalCallbacks[any, gotp.Msg]
	server *gen_server.Server[any, gotp.Msg, T, gotp.Msg, gotp.Msg]
}

func New[T any](fn InitFn[T]) *Agent[T] {
	a := &Agent[T]{initFn: fn}
	a.server = gen_server.New(a, nil)
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

func (a *Agent[T]) Init(any) (gen_server.Continue[gotp.Msg], error) {
	if a.initFn == nil {
		return gen_server.NoCont[gotp.Msg](), fmt.Errorf("init function cannot be nil")
	}
	a.state = a.initFn()
	return gen_server.NoCont[gotp.Msg](), nil
}

func (a *Agent[T]) HandleCall(msg gotp.Msg, from gotp.PID) (gen_server.Response[T], gen_server.Continue[gotp.Msg], error) {
	slog.Debug("HandleCall", slog.String("msg", fmt.Sprintf("%+v", msg)), slog.Any("from", from))
	switch msg := msg.(type) {
	case GetFn[T]:
		return gen_server.Reply(msg(*a.state)), gen_server.NoCont[gotp.Msg](), nil
	case GetAndUpdateFn[T]:
		return gen_server.Reply(msg(a.state)), gen_server.NoCont[gotp.Msg](), nil
	case UpdateFn[T]:
		msg(a.state)
		return gen_server.NoReply[T](), gen_server.NoCont[gotp.Msg](), nil
	default:
		cont, err := a.HandleInfo(msg)
		return gen_server.NoReply[T](), cont, err
	}
}

func (a *Agent[T]) HandleCast(msg gotp.Msg) (gen_server.Continue[gotp.Msg], error) {
	slog.Debug("HandleCast", slog.String("msg", fmt.Sprintf("%+v", msg)))
	switch msg := msg.(type) {
	case UpdateFn[T]:
		msg(a.state)
		return gen_server.NoCont[gotp.Msg](), nil
	default:
		return a.HandleInfo(msg)
	}
}
