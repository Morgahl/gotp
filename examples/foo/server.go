package foo

import (
	"log/slog"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/server"
)

var _ gotp.Supervisable = &FooServer{}
var _ server.Serverable[any, any, any, any, any] = &FooServer{}

type FooServer struct {
	name string

	// Embed the server.DefaultHandlers to provide default implementations
	// for the server.Serverable interface methods.
	server.DefaultHandlers[any]
}

func NewServer(name string) *FooServer {
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

func (f *FooServer) Terminate(reason error) error {
	slog.Debug("FooServer.Terminate called", "name", f.name, "error", reason)
	return reason
}
