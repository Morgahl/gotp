package foo

import (
	"log/slog"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/server"
)

var _ gotp.Supervisable = &FooServer{}
var _ server.Serverable[gotp.Msg, gotp.Msg, gotp.Msg, gotp.Msg, any] = &FooServer{}

type FooServer struct {
	name string
}

func NewServer(name string) gotp.Supervisable {
	return &FooServer{
		name: name,
	}
}

func (f *FooServer) ChildSpec() gotp.ChildSpec {
	return gotp.ChildSpec{
		Name:     f.name,
		Restart:  gotp.PERMANENT,
		Shutdown: 30 * time.Second,
		Type:     gotp.WORKER,
	}
}

func (f *FooServer) Start(timeout time.Duration, opts ...gotp.SpawnOpt) (gotp.Started, error) {
	return server.New(f).Start(timeout, opts...)
}

func (f *FooServer) StartLink(link gotp.PID, timeout time.Duration, opts ...gotp.SpawnOpt) (gotp.Supervised, error) {
	return server.New(f).StartLink(link, timeout, opts...)
}

func (f *FooServer) Init(opts gotp.Options) (server.Continue[any], error) {
	slog.Debug("FooServer.Init called", "name", f.name, "opts", opts)
	return server.NoCont[any](), nil
}

func (f *FooServer) HandleCall(call gotp.Msg, from gotp.PID) (server.Response[gotp.Msg], server.Continue[any], error) {
	return server.NoReply[gotp.Msg](), server.NoCont[any](), nil
}

func (f *FooServer) HandleCast(cast gotp.Msg) (server.Continue[any], error) {
	return server.NoCont[any](), nil
}

func (f *FooServer) HandleContinue(cont any) (server.Continue[any], error) {
	return server.NoCont[any](), nil
}

func (f *FooServer) HandleInfo(info gotp.Msg) (server.Continue[any], error) {
	return server.NoCont[any](), nil
}

func (f *FooServer) HandleAny(msg gotp.Msg) (server.Continue[any], error) {
	return server.NoCont[any](), nil
}

func (f *FooServer) Terminate(err error) error {
	slog.Debug("FooServer.Terminate called", "error", err)
	return err
}
