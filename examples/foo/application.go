package foo

import (
	"context"
	"log"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/application"
	"github.com/Morgahl/gotp/server"
	"github.com/Morgahl/gotp/supervisor"
)

func Start(st application.StartType, args ...any) supervisor.Supervisor {
	srvr := server.New(&NoOp[any, any, any, any, any]{})
	dynamic := supervisor.Dynamic(supervisor.Flags{})
	root := supervisor.Static(supervisor.Flags{}, dynamic, srvr)
	return root
}

var _ server.Serverable[any, any, any, any, any] = &NoOp[any, any, any, any, any]{}

type NoOp[Call any, Resp any, Cast any, Info any, Cont any] struct {
	process *gotp.Process
}

func (f *NoOp[Call, Resp, Cast, Info, Cont]) ID() gotp.PID {
	return f.process.ID()
}

func (f *NoOp[Call, Resp, Cast, Info, Cont]) ChildSpec() supervisor.ChildSpec {
	return supervisor.ChildSpec{
		Restart: supervisor.PERMANENT,
	}
}

func (f *NoOp[Call, Resp, Cast, Info, Cont]) StartLink(ctx context.Context, link gotp.PID, opts ...gotp.SpawnOpt) (supervisor.Supervisable, error) {
	log.Printf("NoOp.StartLink: %v", link)
	return server.New(f).StartLink(ctx, link, opts...)
}

func (f *NoOp[Call, Resp, Cast, Info, Cont]) Init(opts supervisor.Options) (server.Continue[Cont], error) {
	log.Printf("NoOp.Init: %v", opts)
	return server.NoCont[Cont](), nil
}

func (f *NoOp[Call, Resp, Cast, Info, Cont]) HandleCall(call Call, from gotp.PID) (server.Response[Resp], server.Continue[Cont], error) {
	log.Printf("NoOp.HandleCall: %v", call)
	return server.NoReply[Resp](), server.NoCont[Cont](), nil
}

func (f *NoOp[Call, Resp, Cast, Info, Cont]) HandleCast(cast Cast) (server.Continue[Cont], error) {
	log.Printf("NoOp.HandleCast: %v", cast)
	return server.NoCont[Cont](), nil
}

func (f *NoOp[Call, Resp, Cast, Info, Cont]) HandleContinue(cont Cont) (server.Continue[Cont], error) {
	log.Printf("NoOp.HandleContinue: %v", cont)
	return server.NoCont[Cont](), nil
}

func (f *NoOp[Call, Resp, Cast, Info, Cont]) HandleInfo(info Info) (server.Continue[Cont], error) {
	log.Printf("NoOp.HandleInfo: %v", info)
	return server.NoCont[Cont](), nil
}

func (f *NoOp[Call, Resp, Cast, Info, Cont]) HandleAny(msg gotp.Msg) (server.Continue[Cont], error) {
	log.Printf("NoOp.HandleAny: %v", msg)
	return server.NoCont[Cont](), nil
}

func (f *NoOp[Call, Resp, Cast, Info, Cont]) Terminate(err error) error {
	log.Printf("NoOp.Terminate: %v", err)
	return nil
}
