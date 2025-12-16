package agent

import (
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/gen_server"
	"github.com/Morgahl/gotp/process"
	"github.com/Morgahl/gotp/supervisor"
)

func Start[T any](initFn InitFn[T], opts ...process.SpawnOpt) (process.Ref, error) {
	return gen_server.Start(&agent[T]{initFn: initFn}, initFn, opts...)
}

func StartLink[T any](initFn InitFn[T], linked process.Ref, opts ...process.SpawnOpt) (process.Ref, error) {
	return gen_server.StartLink(&agent[T]{initFn: initFn}, initFn, linked, opts...)
}

func Cast[T any, S process.Sendable](to S, msg UpdateFn[T]) {
	gen_server.Cast[gotp.Term](to, msg)
}

func ContextCast[T any, S process.Sendable](pctx process.Context, to S, msg UpdateFn[T]) {
	gen_server.ContextCast[gotp.Term](to, pctx, msg)
}

func Get[T any, S process.Sendable](to S, from process.PID, msg GetFn[T], timeout time.Duration) (T, bool) {
	return gen_server.Call[gotp.Term, T](to, from, msg, timeout)
}

func ContextGet[T any, S process.Sendable](pctx process.Context, to S, msg GetFn[T], timeout time.Duration) (T, bool) {
	return gen_server.ContextCall[gotp.Term, T](to, pctx, msg, timeout)
}

func GetAndUpdate[T any, S process.Sendable](to S, from process.PID, msg GetAndUpdateFn[T], timeout time.Duration) (T, bool) {
	return gen_server.Call[gotp.Term, T](to, from, msg, timeout)
}

func ContextGetAndUpdate[T any, S process.Sendable](pctx process.Context, to S, msg GetAndUpdateFn[T], timeout time.Duration) (T, bool) {
	return gen_server.ContextCall[gotp.Term, T](to, pctx, msg, timeout)
}

func Update[T any, S process.Sendable](to S, msg UpdateFn[T]) {
	gen_server.Cast[gotp.Term](to, msg)
}

func ContextUpdate[T any, S process.Sendable](pctx process.Context, to S, msg UpdateFn[T]) {
	gen_server.ContextCast[gotp.Term](to, pctx, msg)
}

func Stop[S process.Sendable](to S, reason error) {
	process.Send(to, gen_server.StopMsg(reason))
}

func ContextStop[S process.Sendable](pctx process.Context, to S, reason error) {
	process.ContextSend(pctx, to, gen_server.StopMsg(reason))
}

func ChildSpec[T any](InitFn InitFn[T], sopts ...process.SpawnOpt) supervisor.ChildSpec {
	return supervisor.ChildSpec{
		Restart:     supervisor.PERMANENT,
		Shutdown:    gotp.DEFAULT_SHUTDOWN,
		Type:        supervisor.WORKER,
		Significant: true,
		Start: func(opts ...process.SpawnOpt) (process.Ref, error) {
			return gen_server.Start(&agent[T]{initFn: InitFn}, InitFn, append(sopts, opts...)...)
		},
	}
}
