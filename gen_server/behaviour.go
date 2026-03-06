package gen_server

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/Morgahl/gotp/process"
	"github.com/Morgahl/gotp/term"
)

type GenServer[
	I any,
	Cl term.Term,
	R term.Term,
	Cs term.Term,
	Ct term.Term,
] interface {
	Init(process.Context, I) (Continue[Ct], error)
	HandleCall(process.Context, Cl, process.PID) (Response[R], Continue[Ct], error)
	HandleCast(process.Context, Cs) (Continue[Ct], error)
	HandleContinue(process.Context, Ct) (Continue[Ct], error)
	HandleInfo(process.Context, term.Term) (Continue[Ct], error)
	Terminate(process.Context, error) error
}

const (
	DEFAULT_SHUTDOWN = 30 * time.Second
)

// OptionalCallbacks provides default implementations for the Serverable interface methods that log a warning when not
// implemented and are called by the Server implementation. This allows for a server to be created without implementing
// all methods of the Serverable interface, making it easier to create simple servers without boilerplate code.
// It is recommended to use this only for simple servers or during development. For production servers, it is
// recommended to implement all methods of the Serverable interface.
type OptionalCallbacks struct{}

func (OptionalCallbacks) HandleCall(pctx process.Context, msg term.Term, from process.PID) (Response[term.Term], Continue[term.Term], error) {
	slog.WarnContext(pctx.Context(), "HandleCall not implemented, msg will be ignored", slog.String("msg", fmt.Sprintf("%+v", msg)), slog.String("from", from.String()))
	return NoReply[term.Term](), NoCont[term.Term](), nil
}

func (OptionalCallbacks) HandleCast(pctx process.Context, msg term.Term) (Continue[term.Term], error) {
	slog.WarnContext(pctx.Context(), "HandleCast not implemented, msg will be ignored", slog.String("msg", fmt.Sprintf("%+v", msg)))
	return NoCont[term.Term](), nil
}

func (OptionalCallbacks) HandleContinue(pctx process.Context, msg term.Term) (Continue[term.Term], error) {
	slog.WarnContext(pctx.Context(), "HandleContinue not implemented, msg will be ignored", slog.String("msg", fmt.Sprintf("%+v", msg)))
	return NoCont[term.Term](), nil
}

func (OptionalCallbacks) HandleInfo(pctx process.Context, msg term.Term) (Continue[term.Term], error) {
	slog.WarnContext(pctx.Context(), "HandleInfo not implemented, msg will be ignored", slog.String("msg", fmt.Sprintf("%+v", msg)))
	return NoCont[term.Term](), nil
}

func (OptionalCallbacks) Terminate(pctx process.Context, reason error) error {
	slog.WarnContext(pctx.Context(), "Terminate not implemented, server will be stopped", slog.Any("reason", reason))
	return reason
}
