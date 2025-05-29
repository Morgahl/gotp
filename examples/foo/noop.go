package foo

import (
	"log/slog"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/server"
)

var _ server.Serverable[gotp.Msg, gotp.Msg, gotp.Msg, gotp.Msg, any] = &NoOp{}
var _ gotp.Supervisable = &NoOp{}

type NoOp struct {
	*gotp.Process
}

func (f *NoOp) PID() gotp.PID {
	return f.Process.PID()
}

func (f *NoOp) ChildSpec() gotp.ChildSpec {
	return gotp.ChildSpec{
		Restart: gotp.PERMANENT,
	}
}

func (f *NoOp) StartLink(link gotp.PID, timeout time.Duration, opts ...gotp.SpawnOpt) (gotp.Supervisable, error) {
	slog.Debug("NoOp.StartLink", "link", link)
	return server.New(f).StartLink(link, timeout, opts...)
}

func (f *NoOp) Init(opts gotp.Options) (server.Continue[any], error) {
	slog.Debug("NoOp.Init", "opts", opts)
	return server.NoCont[any](), nil
}

func (f *NoOp) HandleCall(call gotp.Msg, from gotp.PID) (server.Response[gotp.Msg], server.Continue[any], error) {
	slog.Debug("NoOp.HandleCall", "call", call)
	return server.NoReply[gotp.Msg](), server.NoCont[any](), nil
}

func (f *NoOp) HandleCast(cast gotp.Msg) (server.Continue[any], error) {
	slog.Debug("NoOp.HandleCast", "cast", cast)
	return server.NoCont[any](), nil
}

func (f *NoOp) HandleContinue(cont any) (server.Continue[any], error) {
	slog.Debug("NoOp.HandleContinue", "cont", cont)
	return server.NoCont[any](), nil
}

func (f *NoOp) HandleInfo(info gotp.Msg) (server.Continue[any], error) {
	slog.Debug("NoOp.HandleInfo", "info", info)
	return server.NoCont[any](), nil
}

func (f *NoOp) HandleAny(msg gotp.Msg) (server.Continue[any], error) {
	slog.Debug("NoOp.HandleAny", "msg", msg)
	return server.NoCont[any](), nil
}

func (f *NoOp) Terminate(err error) error {
	slog.Debug("NoOp.Terminate", "error", err)
	return err
}
