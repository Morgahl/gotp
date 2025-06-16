package agent

import (
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/server"
)

var _ server.Serverable[any, any, any, any] = &Agent[any]{}

type InitFn[T any] func() *T

type GetFn[T any] func(T) T

type GetAndUpdateFn[T any] func(*T) T

type UpdateFn[T any] func(*T)

type Agent[T any] struct {
	initFn InitFn[T]
	state  *T

	server.DefaultHandlers[any]
	server *server.Server[any, T, any, any]
}

func New[Fn InitFn[T], T any](fn Fn) *Agent[T] {
	a := &Agent[T]{}
	a.server = server.New(a)
	return a
}

func (a *Agent[T]) ChildSpec() gotp.ChildSpec {
	return gotp.ChildSpec{
		Restart:     gotp.PERMANENT,
		Shutdown:    30 * time.Second,
		Type:        gotp.WORKER,
		Significant: true,
	}
}

func (a *Agent[T]) Start(opts ...gotp.SpawnOpt) (gotp.Started, error) {
	return a.server.Start(opts...)
}

func (a *Agent[T]) StartLink(link gotp.PID, opts ...gotp.SpawnOpt) (gotp.Supervised, error) {
	return a.server.StartLink(link, opts...)
}

func (a *Agent[T]) Init(gotp.Options) (server.Continue[any], error) {
	a.state = a.initFn()
	return server.NoCont[any](), nil
}

func (a *Agent[T]) HandleCall(msg any, from gotp.PID) (server.Response[T], server.Continue[any], error) {
	switch msg := msg.(type) {
	case GetFn[T]:
		return server.Reply(msg(*a.state)), server.NoCont[any](), nil
	case GetAndUpdateFn[T]:
		return server.Reply(msg(a.state)), server.NoCont[any](), nil
	case UpdateFn[T]:
		msg(a.state)
		return server.NoReply[T](), server.NoCont[any](), nil
	default:
		cont, err := a.HandleInfo(msg)
		return server.NoReply[T](), cont, err
	}
}

func (a *Agent[T]) HandleCast(msg any) (server.Continue[any], error) {
	switch msg := msg.(type) {
	case UpdateFn[T]:
		msg(a.state)
		return server.NoCont[any](), nil
	default:
		return a.HandleInfo(msg)
	}
}
