package server

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/process"
	"github.com/Morgahl/gotp/supervisor"
)

type Serverable[
	I any,
	Cl process.Message,
	R process.Message,
	Cs process.Message,
	Ct process.Message,
] interface {
	ChildSpec() supervisor.ChildSpec
	Init(I) (Continue[Ct], error)
	HandleCall(Cl, process.PID) (Response[R], Continue[Ct], error)
	HandleCast(Cs) (Continue[Ct], error)
	HandleContinue(Ct) (Continue[Ct], error)
	HandleInfo(process.Message) (Continue[Ct], error)
	Terminate(error) error
}

const (
	DEFAULT_SHUTDOWN = 30 * time.Second
)

// OptionalCallbacks provides default implementations for the Serverable interface methods that log a warning when not
// implemented and are called by the Server implementation. This allows for a server to be created without implementing
// all methods of the Serverable interface, making it easier to create simple servers without boilerplate code.
// It is recommended to use this only for simple servers or during development. For production servers, it is
// recommended to implement all methods of the Serverable interface.
//
// NOTE: This expects two generic type parameters that the implementor must provide the correct type for it. This is to
// ensure that the server correctly matches the Serverable interface:
// - I which is the arg passed to Init when it is called
// - Ct which is the type of the continue message
type OptionalCallbacks[I any, Ct process.Message] struct{}

func (OptionalCallbacks[I, Ct]) ChildSpec() supervisor.ChildSpec {
	slog.Warn("ChildSpec not implemented, defaulting to permanent worker")
	return supervisor.ChildSpec{
		Restart:  supervisor.PERMANENT,
		Shutdown: gotp.DEFAULT_SHUTDOWN,
		Type:     supervisor.WORKER,
	}
}

func (OptionalCallbacks[I, Ct]) HandleCall(msg any, from process.PID) (Response[any], Continue[Ct], error) {
	slog.Warn("HandleCall not implemented, msg will be ignored", slog.String("msg", fmt.Sprintf("%+v", msg)), slog.String("from", from.String()))
	return NoReply[any](), NoCont[Ct](), nil
}

func (OptionalCallbacks[I, Ct]) HandleCast(msg any) (Continue[Ct], error) {
	slog.Warn("HandleCast not implemented, msg will be ignored", slog.String("msg", fmt.Sprintf("%+v", msg)))
	return NoCont[Ct](), nil
}

func (OptionalCallbacks[I, Ct]) HandleContinue(msg Ct) (Continue[Ct], error) {
	slog.Warn("HandleContinue not implemented, msg will be ignored", slog.String("msg", fmt.Sprintf("%+v", msg)))
	return NoCont[Ct](), nil
}

func (OptionalCallbacks[I, Ct]) HandleInfo(msg process.Message) (Continue[Ct], error) {
	slog.Warn("HandleInfo not implemented, msg will be ignored", slog.String("msg", fmt.Sprintf("%+v", msg)))
	return NoCont[Ct](), nil
}

func (OptionalCallbacks[I, Ct]) Terminate(reason error) error {
	slog.Warn("Terminate not implemented, server will be stopped", slog.Any("reason", reason))
	return reason
}
