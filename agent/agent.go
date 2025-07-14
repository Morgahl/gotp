package agent

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/process"
	"github.com/Morgahl/gotp/server"
)

var _ server.Serverable[InitFn[int], process.Message, int, process.Message, process.Message] = &Agent[int]{}

type InitFn[T any] func() *T

type GetFn[T any] func(T) T

type GetAndUpdateFn[T any] func(*T) T

type UpdateFn[T any] func(*T)

type Agent[T any] struct {
	state  *T
	server *server.Server[InitFn[T], process.Message, T, process.Message, process.Message]
	server.OptionalCallbacks[any, process.Message]
}

func New[T any](fn InitFn[T]) *Agent[T] {
	a := &Agent[T]{}
	a.server = server.New(a, fn)
	return a
}

func (a *Agent[T]) Start(opts ...process.SpawnOpt) (process.Started, error) {
	return a.server.Start(opts...)
}

func (a *Agent[T]) StartLink(linked *process.Process, opts ...process.SpawnOpt) (server.Supervised, error) {
	return a.server.StartLink(linked, opts...)
}

func (a *Agent[T]) Context() context.Context {
	return a.server.Process().Context()
}

func (a *Agent[T]) ChildSpec() server.ChildSpec {
	return server.ChildSpec{
		Restart:     server.PERMANENT,
		Shutdown:    gotp.DEFAULT_SHUTDOWN,
		Type:        server.WORKER,
		Significant: true,
	}
}

func (a *Agent[T]) Init(initFn InitFn[T]) (server.Continue[process.Message], error) {
	a.state = initFn()
	return server.NoCont[process.Message](), nil
}

func (a *Agent[T]) HandleCall(msg process.Message, from process.PID) (server.Response[T], server.Continue[process.Message], error) {
	switch msg := msg.(type) {
	case GetFn[T]:
		slog.DebugContext(a.Context(), "HandleCall", slog.Any("from", from), slog.Any("msg", fmt.Sprintf("%T", msg)))
		return server.Reply(msg(*a.state)), server.NoCont[process.Message](), nil
	case GetAndUpdateFn[T]:
		slog.DebugContext(a.Context(), "HandleCall", slog.Any("from", from), slog.Any("msg", fmt.Sprintf("%T", msg)))
		return server.Reply(msg(a.state)), server.NoCont[process.Message](), nil
	case UpdateFn[T]:
		slog.DebugContext(a.Context(), "HandleCall", slog.Any("from", from), slog.Any("msg", fmt.Sprintf("%T", msg)))
		msg(a.state)
		return server.NoReply[T](), server.NoCont[process.Message](), nil
	default:
		slog.DebugContext(a.Context(), "HandleCall", slog.Any("from", from), slog.Any("msg", fmt.Sprintf("%T", msg)))
		cont, err := a.HandleInfo(msg)
		return server.NoReply[T](), cont, err
	}
}

func (a *Agent[T]) HandleCast(msg process.Message) (server.Continue[process.Message], error) {
	switch msg := msg.(type) {
	case UpdateFn[T]:
		slog.DebugContext(a.Context(), "HandleCast", slog.Any("msg", fmt.Sprintf("%T", msg)))
		msg(a.state)
		return server.NoCont[process.Message](), nil
	default:
		return a.HandleInfo(msg)
	}
}
