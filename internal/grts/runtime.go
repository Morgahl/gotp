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

func Run(app application.Application) (err error) {
	ctx := ctx.Root()
	defer ctx.Cancel(fmt.Errorf("grts.Run: exiting"))

	wg := &sync.WaitGroup{}
	wg.Add(1)
	root := process.Spawn(_init(ctx, app, wg))
	root.Start()
	slog.DebugContext(ctx, "grts.Run: root process started", slog.Any("pid", root.PID()))

	<-ctx.Done()
	wg.Wait()

	slog.Info("grts.Run: exiting")
	return context.Cause(ctx)
}

func _init(rootCtx ctx.Cancellable, app application.Application, wg *sync.WaitGroup) process.RunFn {
	return func(p *process.Process) (reason error) {
		defer wg.Done()
		defer func() {
			reason = debug.Recover(recover(), "grts._init", reason)
			rootCtx.Cancel(reason)
		}()

		slog.DebugContext(rootCtx, "grts._init: starting application", slog.Any("name", app.Name()), slog.Any("version", app.Version()))

		var root server.Supervisable
		startUp := time.Now()
		if root, reason = app.Start(application.Normal()); reason != nil {
			slog.ErrorContext(rootCtx, "grts._init: failed to call application start hook", slog.Any("error", reason))
			return reason
		}

		slog.DebugContext(rootCtx, "grts._init: application start hook called", slog.Duration("took", time.Since(startUp)))
		var sup server.Supervised
		if sup, reason = root.StartLink(p); reason != nil {
			slog.ErrorContext(rootCtx, "grts._init: failed to start supervision tree", slog.Any("error", reason))
			return reason
		}

		appStart := time.Now()
		defer func(appStart time.Time) {
			slog.InfoContext(rootCtx, "grts._init: application exited", slog.Duration("after", time.Since(appStart)))
		}(appStart)
		slog.DebugContext(rootCtx, "grts._init: supervision tree started", slog.Any("pid", sup.PID()), slog.Duration("took", time.Since(startUp)))

		if msg, ok, err := process.ReceiveContext[process.ExitMsg](p, rootCtx); err != nil {
			if cause := context.Cause(rootCtx); errors.Is(err, cause) {
				slog.DebugContext(rootCtx, "grts._init: received expected shutdown signal", slog.String("reason", err.Error()))
				reason = cause
				goto EXIT
			}
			slog.ErrorContext(rootCtx, "grts._init: unexpected error receiving ExitMsg", slog.Any("pid", p.PID()), slog.Any("error", err))
			reason = err
			goto EXIT
		} else if ok {
			slog.DebugContext(rootCtx, "grts._init: received exit", slog.String("msg", fmt.Sprintf("%+v", msg)))
			reason = msg.Reason
			goto EXIT
		}

	EXIT:
		slog.DebugContext(rootCtx, "grts._init: shutting down supervision tree")
		shutDown := time.Now()
		sup.Exit(reason)
		slog.DebugContext(rootCtx, "grts._init: waiting for exit messages", slog.Duration("took", time.Since(shutDown)))
		msg, ok, err := process.ReceiveWithTimeout[process.ExitMsg](p, 0)
		if err != nil {
			if errors.Is(err, reason) {
				slog.DebugContext(rootCtx, "grts._init: received expected exit message", slog.Any("pid", p.PID()), slog.Any("reason", reason))
				return reason
			}
			slog.ErrorContext(rootCtx, "grts._init: error receiving ExitMsg", slog.Any("pid", p.PID()), slog.Any("error", err))
			return err
		} else if !ok {
			slog.DebugContext(rootCtx, "grts._init: process exited without exit message", slog.Duration("took", time.Since(shutDown)))
		}
		slog.InfoContext(rootCtx, "grts._init: exiting", slog.Duration("took", time.Since(shutDown)))
		return msg.Reason
	}
}
