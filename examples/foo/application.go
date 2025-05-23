package foo

import (
	"log/slog"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/application"
	"github.com/Morgahl/gotp/server"
	"github.com/Morgahl/gotp/supervisor"
)

func Start(st application.StartType, args ...any) supervisor.Supervisor {
	srvr := server.New(&NoOp[any, any, any, any, any]{})
	static := supervisor.Static(supervisor.Flags{}, srvr)
	dynamic := supervisor.Dynamic(supervisor.Flags{})
	root := supervisor.Static(supervisor.Flags{}, static, dynamic)
	return root
}

var _ server.Serverable[any, any, any, any, any] = &NoOp[any, any, any, any, any]{}

type NoOp[Call any, Resp any, Cast any, Info any, Cont any] struct {
	*gotp.Process
}

func (f *NoOp[Call, Resp, Cast, Info, Cont]) ID() gotp.PID {
	return f.Process.ID()
}

func (f *NoOp[Call, Resp, Cast, Info, Cont]) ChildSpec() supervisor.ChildSpec {
	return supervisor.ChildSpec{
		Restart: supervisor.PERMANENT,
	}
}

func (f *NoOp[Call, Resp, Cast, Info, Cont]) StartLink(link gotp.PID, timeout time.Duration, opts ...gotp.SpawnOpt) (supervisor.Supervisable, error) {
	slog.Debug("NoOp.StartLink", "link", link)
	return server.New(f).StartLink(link, timeout, opts...)
}

func (f *NoOp[Call, Resp, Cast, Info, Cont]) Init(opts supervisor.Options) (server.Continue[Cont], error) {
	slog.Debug("NoOp.Init", "opts", opts)
	return server.NoCont[Cont](), nil
}

func (f *NoOp[Call, Resp, Cast, Info, Cont]) HandleCall(call Call, from gotp.PID) (server.Response[Resp], server.Continue[Cont], error) {
	slog.Debug("NoOp.HandleCall", "call", call)
	return server.NoReply[Resp](), server.NoCont[Cont](), nil
}

func (f *NoOp[Call, Resp, Cast, Info, Cont]) HandleCast(cast Cast) (server.Continue[Cont], error) {
	slog.Debug("NoOp.HandleCast", "cast", cast)
	return server.NoCont[Cont](), nil
}

func (f *NoOp[Call, Resp, Cast, Info, Cont]) HandleContinue(cont Cont) (server.Continue[Cont], error) {
	slog.Debug("NoOp.HandleContinue", "cont", cont)
	return server.NoCont[Cont](), nil
}

func (f *NoOp[Call, Resp, Cast, Info, Cont]) HandleInfo(info Info) (server.Continue[Cont], error) {
	slog.Debug("NoOp.HandleInfo", "info", info)
	return server.NoCont[Cont](), nil
}

func (f *NoOp[Call, Resp, Cast, Info, Cont]) HandleAny(msg gotp.Msg) (server.Continue[Cont], error) {
	slog.Debug("NoOp.HandleAny", "msg", msg)
	return server.NoCont[Cont](), nil
}

func (f *NoOp[Call, Resp, Cast, Info, Cont]) Terminate(err error) error {
	slog.Debug("NoOp.Terminate", "error", err)
	return err
}
