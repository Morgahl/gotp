package gen_server

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/process"
)

type GenServer[
	I any,
	Cl gotp.Term,
	R gotp.Term,
	Cs gotp.Term,
	Ct gotp.Term,
] interface {
	Init(process.Context, I) (Continue[Ct], error)
	HandleCall(process.Context, Cl, process.PID) (Response[R], Continue[Ct], error)
	HandleCast(process.Context, Cs) (Continue[Ct], error)
	HandleContinue(process.Context, Ct) (Continue[Ct], error)
	HandleInfo(process.Context, gotp.Term) (Continue[Ct], error)
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

func (OptionalCallbacks) HandleCall(pctx process.Context, msg gotp.Term, from process.PID) (Response[gotp.Term], Continue[gotp.Term], error) {
	slog.WarnContext(pctx.Context(), "HandleCall not implemented, msg will be ignored", slog.String("msg", fmt.Sprintf("%+v", msg)), slog.String("from", from.String()))
	return NoReply[gotp.Term](), NoCont[gotp.Term](), nil
}

func (OptionalCallbacks) HandleCast(pctx process.Context, msg gotp.Term) (Continue[gotp.Term], error) {
	slog.WarnContext(pctx.Context(), "HandleCast not implemented, msg will be ignored", slog.String("msg", fmt.Sprintf("%+v", msg)))
	return NoCont[gotp.Term](), nil
}

func (OptionalCallbacks) HandleContinue(pctx process.Context, msg gotp.Term) (Continue[gotp.Term], error) {
	slog.WarnContext(pctx.Context(), "HandleContinue not implemented, msg will be ignored", slog.String("msg", fmt.Sprintf("%+v", msg)))
	return NoCont[gotp.Term](), nil
}

func (OptionalCallbacks) HandleInfo(pctx process.Context, msg gotp.Term) (Continue[gotp.Term], error) {
	slog.WarnContext(pctx.Context(), "HandleInfo not implemented, msg will be ignored", slog.String("msg", fmt.Sprintf("%+v", msg)))
	return NoCont[gotp.Term](), nil
}

func (OptionalCallbacks) Terminate(pctx process.Context, reason error) error {
	slog.WarnContext(pctx.Context(), "Terminate not implemented, server will be stopped", slog.Any("reason", reason))
	return reason
}
