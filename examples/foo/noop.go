package foo

import (
	"log/slog"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/server"
)

var _ server.Serverable[gotp.Msg, gotp.Msg, gotp.Msg, gotp.Msg, any] = &NoOp[gotp.Msg, gotp.Msg, gotp.Msg, gotp.Msg, any]{}
var _ gotp.Supervisable = &NoOp[gotp.Msg, gotp.Msg, gotp.Msg, gotp.Msg, any]{}

type NoOp[Call any, Resp any, Cast any, Info any, Cont any] struct {
	*gotp.Process
}

func (f *NoOp[Call, Resp, Cast, Info, Cont]) PID() gotp.PID {
	return f.Process.PID()
}

func (f *NoOp[Call, Resp, Cast, Info, Cont]) ChildSpec() gotp.ChildSpec {
	return gotp.ChildSpec{
		Restart: gotp.PERMANENT,
	}
}

func (f *NoOp[Call, Resp, Cast, Info, Cont]) StartLink(link gotp.PID, timeout time.Duration, opts ...gotp.SpawnOpt) (gotp.Supervisable, error) {
	slog.Debug("NoOp.StartLink", "link", link)
	return server.New(f).StartLink(link, timeout, opts...)
}

func (f *NoOp[Call, Resp, Cast, Info, Cont]) Init(opts gotp.Options) (server.Continue[Cont], error) {
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
