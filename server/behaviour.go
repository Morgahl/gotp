package server

import (
	"log/slog"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/debug"
)

type Serverable[
	I any,
	Cl gotp.Msg,
	R gotp.Msg,
	Cs gotp.Msg,
	Ct gotp.Msg,
] interface {
	ChildSpec() gotp.ChildSpec
	Init(I) (Continue[Ct], error)
	HandleCall(Cl, gotp.PID) (Response[R], Continue[Ct], error)
	HandleCast(Cs) (Continue[Ct], error)
	HandleContinue(Ct) (Continue[Ct], error)
	HandleInfo(gotp.Msg) (Continue[Ct], error)
	Terminate(error) error
}

const (
	DEFAULT_SHUTDOWN = 5 * time.Second
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
type OptionalCallbacks[I any, Ct gotp.Msg] struct{}

func (OptionalCallbacks[I, Ct]) ChildSpec() gotp.ChildSpec {
	slog.Warn("ChildSpec not implemented, defaulting to permanent worker")
	return gotp.ChildSpec{
		Restart:  gotp.PERMANENT,
		Shutdown: DEFAULT_SHUTDOWN,
		Type:     gotp.WORKER,
	}
}

func (OptionalCallbacks[I, Ct]) HandleCall(call any, from gotp.PID) (Response[any], Continue[Ct], error) {
	slog.Warn("HandleCall not implemented, call will be ignored", slog.Any("call", call), slog.String("from", from.String()))
	return NoReply[any](), NoCont[Ct](), nil
}

func (OptionalCallbacks[I, Ct]) HandleCast(cast any) (Continue[Ct], error) {
	slog.Warn("HandleCast not implemented, cast will be ignored", slog.Any("cast", cast))
	return NoCont[Ct](), nil
}

func (OptionalCallbacks[I, Ct]) HandleContinue(cont Ct) (Continue[Ct], error) {
	slog.Warn("HandleContinue not implemented, continue will be ignored", slog.Any("continue", cont))
	return NoCont[Ct](), nil
}

func (OptionalCallbacks[I, Ct]) HandleInfo(info gotp.Msg) (Continue[Ct], error) {
	slog.Warn("HandleInfo not implemented, info will be ignored", slog.Any("info", info))
	panic(debug.ThrowF("HandleInfo not implemented, info will be ignored: %T", info))
	return NoCont[Ct](), nil
}

func (OptionalCallbacks[I, Ct]) Terminate(reason error) error {
	slog.Warn("Terminate not implemented, server will be stopped", slog.String("reason", reason.Error()))
	return reason
}
