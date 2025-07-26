package gen_server

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/Morgahl/gotp/process"
)

type GenServer[
	I any,
	Cl process.Message,
	R process.Message,
	Cs process.Message,
	Ct process.Message,
] interface {
	Init(*process.Context, I) (Continue[Ct], error)
	HandleCall(*process.Context, Cl, process.PID) (Response[R], Continue[Ct], error)
	HandleCast(*process.Context, Cs) (Continue[Ct], error)
	HandleContinue(*process.Context, Ct) (Continue[Ct], error)
	HandleInfo(*process.Context, process.Message) (Continue[Ct], error)
	Terminate(*process.Context, error) error
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

func (OptionalCallbacks[I, Ct]) HandleCall(pctx *process.Context, msg any, from process.PID) (Response[any], Continue[Ct], error) {
	slog.WarnContext(pctx.Context(), "HandleCall not implemented, msg will be ignored", slog.String("msg", fmt.Sprintf("%+v", msg)), slog.String("from", from.String()))
	return NoReply[any](), NoCont[Ct](), nil
}

func (OptionalCallbacks[I, Ct]) HandleCast(pctx *process.Context, msg any) (Continue[Ct], error) {
	slog.WarnContext(pctx.Context(), "HandleCast not implemented, msg will be ignored", slog.String("msg", fmt.Sprintf("%+v", msg)))
	return NoCont[Ct](), nil
}

func (OptionalCallbacks[I, Ct]) HandleContinue(pctx *process.Context, msg Ct) (Continue[Ct], error) {
	slog.WarnContext(pctx.Context(), "HandleContinue not implemented, msg will be ignored", slog.String("msg", fmt.Sprintf("%+v", msg)))
	return NoCont[Ct](), nil
}

func (OptionalCallbacks[I, Ct]) HandleInfo(pctx *process.Context, msg process.Message) (Continue[Ct], error) {
	slog.WarnContext(pctx.Context(), "HandleInfo not implemented, msg will be ignored", slog.String("msg", fmt.Sprintf("%+v", msg)))
	return NoCont[Ct](), nil
}

func (OptionalCallbacks[I, Ct]) Terminate(pctx *process.Context, reason error) error {
	slog.WarnContext(pctx.Context(), "Terminate not implemented, server will be stopped", slog.Any("reason", reason))
	return reason
}
