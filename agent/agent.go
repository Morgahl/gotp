package agent

import (
	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/gen_server"
	"github.com/Morgahl/gotp/process"
	"github.com/Morgahl/gotp/supervisor"
	"github.com/Morgahl/gotp/term"
)

var _ gen_server.GenServer[InitFn[int], term.Term, int, term.Term, term.Term] = &agent[int]{}

type InitFn[T any] func() *T

type GetFn[T any] func(T) T

type GetAndUpdateFn[T any] func(*T) T

type UpdateFn[T any] func(*T)

type agent[T any] struct {
	initFn InitFn[T]
	state  *T
	gen_server.OptionalCallbacks
}

func (a *agent[T]) ChildSpec() supervisor.ChildSpec {
	return supervisor.ChildSpec{
		Restart:  supervisor.PERMANENT,
		Shutdown: gotp.DEFAULT_SHUTDOWN,
		Type:     supervisor.WORKER,
		// Significant: true,
		Start: func(opts ...process.SpawnOpt) (process.Ref, error) {
			return gen_server.Start(a, a.initFn, opts...)
		},
	}
}

func (a *agent[T]) Init(pctx process.Context, initFn InitFn[T]) (gen_server.Continue[term.Term], error) {
	a.state = initFn()
	return gen_server.NoCont[term.Term](), nil
}

func (a *agent[T]) HandleCall(pctx process.Context, msg term.Term, from process.PID) (gen_server.Response[T], gen_server.Continue[term.Term], error) {
	switch msg := msg.(type) {
	case GetFn[T]:
		return gen_server.Reply(msg(*a.state)), gen_server.NoCont[term.Term](), nil
	case GetAndUpdateFn[T]:
		return gen_server.Reply(msg(a.state)), gen_server.NoCont[term.Term](), nil
	case UpdateFn[T]:
		msg(a.state)
		return gen_server.NoReply[T](), gen_server.NoCont[term.Term](), nil
	default:
		cont, err := a.HandleInfo(pctx, msg)
		return gen_server.NoReply[T](), cont, err
	}
}

func (a *agent[T]) HandleCast(pctx process.Context, msg term.Term) (gen_server.Continue[term.Term], error) {
	switch msg := msg.(type) {
	case UpdateFn[T]:
		msg(a.state)
		return gen_server.NoCont[term.Term](), nil
	default:
		return a.HandleInfo(pctx, msg)
	}
}

func (a *agent[T]) HandleInfo(pctx process.Context, msg term.Term) (gen_server.Continue[term.Term], error) {
	switch msg := msg.(type) {
	case process.ExitMsg:
		if msg.PID == pctx.PID() {
			return gen_server.Stop[term.Term](msg.Reason), nil
		}
	}
	return gen_server.NoCont[term.Term](), nil
}

func (a *agent[T]) Terminate(pctx process.Context, reason error) error {
	return reason
}
