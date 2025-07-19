package grts

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/Morgahl/gotp/application"
	"github.com/Morgahl/gotp/debug"
	"github.com/Morgahl/gotp/internal/ctx"
	"github.com/Morgahl/gotp/process"
	"github.com/Morgahl/gotp/supervisor"
)

var (
	init_ref process.Ref
)

func Run(app application.Application) (err error) {
	ctx := ctx.Root()
	defer ctx.Cancel(fmt.Errorf("grts.Run: exiting"))

	wg := &sync.WaitGroup{}
	wg.Add(1)
	init_ref = process.Spawn(_init(ctx, app, wg))
	slog.Debug("grts.Run: init process started", slog.Any("pid", init_ref.PID()))

	<-ctx.Done()
	wg.Wait()

	slog.Debug("grts.Run: exiting")
	return context.Cause(ctx)
}

func Stop(reason error) {
	slog.Debug("grts.Stop: stopping application", slog.Any("reason", reason))
	process.Send(init_ref, process.ExitMsg{PID: init_ref.PID(), Reason: reason})
}

func _init(initCtx ctx.Cancellable, app application.Application, wg *sync.WaitGroup) process.RunFn {
	return func(pctx process.Context) (reason error) {
		pctx.TrapExit(true)
		defer wg.Done()
		defer func() {
			reason = debug.Recover(recover(), "grts._init", reason)
			initCtx.Cancel(reason)
		}()

		slog.DebugContext(pctx.Context(), "grts._init: calling start hook", slog.Any("name", app.Name()), slog.Any("version", app.Version()))
		var init supervisor.Supervisable
		startUp := time.Now()
		if init, reason = app.Start(application.Normal()); reason != nil {
			slog.ErrorContext(pctx.Context(), "grts._init: failed to call application start hook", slog.Any("error", reason))
			return reason
		}

		slog.InfoContext(pctx.Context(), "grts._init: starting application", slog.Duration("took", time.Since(startUp)))
		var sup process.Ref
		if sup, reason = init.ChildSpec().Start(); reason != nil {
			slog.ErrorContext(pctx.Context(), "grts._init: failed to start supervision tree", slog.Any("error", reason))
			return reason
		}

		defer func(startUp time.Time) {
			slog.InfoContext(pctx.Context(), "grts._init: exiting", slog.Duration("after", time.Since(startUp)))
		}(startUp)
		slog.InfoContext(pctx.Context(), "grts._init: supervision tree started", slog.Any("pid", sup.PID()), slog.Duration("took", time.Since(startUp)))
		for {
			if msg, ok, err := process.ReceiveContext[process.ExitMsg](pctx, initCtx); err != nil {
				if cause := context.Cause(initCtx); errors.Is(err, cause) {
					slog.DebugContext(pctx.Context(), "grts._init: received shutdown signal", slog.Any("reason", err))
					reason = cause
					goto EXIT
				}
				slog.ErrorContext(pctx.Context(), "grts._init: unexpected error receiving ExitMsg", slog.Any("error", err))
				reason = err
				goto EXIT
			} else if ok && msg.PID == pctx.PID() {
				slog.InfoContext(pctx.Context(), "grts._init: received exit", slog.Any("msg", msg))
				reason = msg.Reason
				goto EXIT
			}
		}

	EXIT:
		slog.InfoContext(pctx.Context(), "grts._init: shutting down supervision tree", slog.Any("reason", reason))
		shutDown := time.Now()
		defer func(shutDown time.Time) {
			slog.InfoContext(pctx.Context(), "grts._init: application exited", slog.Duration("took", time.Since(shutDown)))
		}(shutDown)
		sup.Send(process.ExitMsg{PID: sup.PID(), Reason: reason})
		slog.DebugContext(pctx.Context(), "grts._init: waiting for exit messages")
		msg, ok, err := process.ReceiveWithTimeout[process.ExitMsg](pctx, 0)
		if err != nil {
			if errors.Is(err, reason) {
				slog.DebugContext(pctx.Context(), "grts._init: received expected exit message", slog.Any("reason", reason))
				return reason
			}
			slog.ErrorContext(pctx.Context(), "grts._init: error receiving ExitMsg", slog.Any("error", err))
			return err
		} else if !ok {
			slog.WarnContext(pctx.Context(), "grts._init: process exited without exit message", slog.Duration("took", time.Since(shutDown)))
		}
		return msg.Reason
	}
}
