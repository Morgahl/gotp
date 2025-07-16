package agent

import (
	"context"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/process"
	"github.com/Morgahl/gotp/server"
)

var _ server.Serverable[InitFn[int], process.Message, int, process.Message, any] = &Agent[int]{}

type InitFn[T any] func() *T

type GetFn[T any] func(T) T

type GetAndUpdateFn[T any] func(*T) T

type UpdateFn[T any] func(*T)

type Agent[T any] struct {
	state  *T
	server *server.Server[InitFn[T], process.Message, T, process.Message, any]
	server.OptionalCallbacks[any, any]
	spawnOpts []process.SpawnOpt
}

func New[T any](fn InitFn[T], opts ...process.SpawnOpt) *Agent[T] {
	a := &Agent[T]{
		spawnOpts: opts,
	}
	a.server = server.New(a, fn)
	return a
}

func (a *Agent[T]) Start(opts ...process.SpawnOpt) (process.Started, error) {
	return a.server.Start(append(a.spawnOpts, opts...)...)
}

func (a *Agent[T]) StartLink(linked *process.Process, opts ...process.SpawnOpt) (server.Supervised, error) {
	return a.server.StartLink(linked, append(a.spawnOpts, opts...)...)
}

func (a *Agent[T]) Context() context.Context {
	return a.server.Process().Context()
}

func (a *Agent[T]) ChildSpec() server.ChildSpec {
	return server.ChildSpec{
		Restart:  server.PERMANENT,
		Shutdown: gotp.DEFAULT_SHUTDOWN,
		Type:     server.WORKER,
		// Significant: true,
	}
}

func (a *Agent[T]) Init(initFn InitFn[T]) (server.Continue[any], error) {
	a.state = initFn()
	return server.NoCont[any](), nil
}

func (a *Agent[T]) HandleCall(msg process.Message, from process.PID) (server.Response[T], server.Continue[any], error) {
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

func (a *Agent[T]) HandleCast(msg process.Message) (server.Continue[any], error) {
	switch msg := msg.(type) {
	case UpdateFn[T]:
		msg(a.state)
		return server.NoCont[any](), nil
	default:
		return a.HandleInfo(msg)
	}
}

func (a *Agent[T]) HandleInfo(msg process.Message) (server.Continue[any], error) {
	switch msg := msg.(type) {
	case process.ExitMsg:
		if msg.PID == a.server.PID() {
			return server.Stop[any](msg.Reason), nil
		}
	}
	return server.NoCont[any](), nil
}

func (a *Agent[T]) Terminate(reason error) error {
	return reason
}
