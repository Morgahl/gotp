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
	"github.com/Morgahl/gotp/server"
)

var (
	init_process *process.Process
)

func Run(app application.Application) (err error) {
	ctx := ctx.Root()
	defer ctx.Cancel(fmt.Errorf("grts.Run: exiting"))

	wg := &sync.WaitGroup{}
	wg.Add(1)
	init_process = process.Spawn(_init(ctx, app, wg))
	init_process.Start()
	slog.DebugContext(init_process.Context(), "grts.Run: init process started")

	<-ctx.Done()
	wg.Wait()

	slog.DebugContext(init_process.Context(), "grts.Run: exiting")
	return context.Cause(ctx)
}

func Stop(reason error) {
	slog.DebugContext(init_process.Context(), "grts.Stop: stopping application", slog.Any("reason", reason))
	process.Send(init_process, process.ExitMsg{PID: init_process.PID(), Reason: reason})
}

func _init(initCtx ctx.Cancellable, app application.Application, wg *sync.WaitGroup) process.RunFn {
	return func(p *process.Process) (reason error) {
		p.UpdateFlags(func(flags process.ProcessFlags) process.ProcessFlags { return flags | process.TRAP_EXIT_FLAG })
		defer wg.Done()
		defer func() {
			reason = debug.Recover(recover(), "grts._init", reason)
			initCtx.Cancel(reason)
		}()

		slog.DebugContext(p.Context(), "grts._init: calling start hook", slog.Any("name", app.Name()), slog.Any("version", app.Version()))

		var init server.Supervisable
		startUp := time.Now()
		if init, reason = app.Start(application.Normal()); reason != nil {
			slog.ErrorContext(p.Context(), "grts._init: failed to call application start hook", slog.Any("error", reason))
			return reason
		}

		slog.InfoContext(p.Context(), "grts._init: starting application", slog.Duration("took", time.Since(startUp)))
		var sup server.Supervised
		if sup, reason = init.StartLink(p, init.ChildSpec().SpawnOpts...); reason != nil {
			slog.ErrorContext(p.Context(), "grts._init: failed to start supervision tree", slog.Any("error", reason))
			return reason
		}

		defer func(startUp time.Time) {
			slog.InfoContext(p.Context(), "grts._init: exiting", slog.Duration("after", time.Since(startUp)))
		}(startUp)
		slog.InfoContext(p.Context(), "grts._init: supervision tree started", slog.Any("pid", sup.PID()), slog.Duration("took", time.Since(startUp)))
		for {
			if msg, ok, err := process.ReceiveContext[process.ExitMsg](p, initCtx); err != nil {
				if cause := context.Cause(initCtx); errors.Is(err, cause) {
					slog.DebugContext(p.Context(), "grts._init: received expected shutdown signal", slog.Any("reason", err))
					reason = cause
					goto EXIT
				}
				slog.ErrorContext(p.Context(), "grts._init: unexpected error receiving ExitMsg", slog.Any("error", err))
				reason = err
				goto EXIT
			} else if ok && msg.PID == p.PID() {
				slog.InfoContext(p.Context(), "grts._init: received exit", slog.Any("msg", msg))
				reason = msg.Reason
				goto EXIT
			}
		}

	EXIT:
		slog.InfoContext(p.Context(), "grts._init: shutting down supervision tree")
		shutDown := time.Now()
		defer func(shutDown time.Time) {
			slog.InfoContext(p.Context(), "grts._init: application exited", slog.Duration("took", time.Since(shutDown)))
		}(shutDown)
		sup.Exit(reason)
		slog.DebugContext(p.Context(), "grts._init: waiting for exit messages", slog.Duration("took", time.Since(shutDown)))
		msg, ok, err := process.ReceiveWithTimeout[process.ExitMsg](p, 0)
		if err != nil {
			if errors.Is(err, reason) {
				slog.DebugContext(p.Context(), "grts._init: received expected exit message", slog.Any("reason", reason))
				return reason
			}
			slog.ErrorContext(p.Context(), "grts._init: error receiving ExitMsg", slog.Any("error", err))
			return err
		} else if !ok {
			slog.WarnContext(p.Context(), "grts._init: process exited without exit message", slog.Duration("took", time.Since(shutDown)))
		}
		return msg.Reason
	}
}
