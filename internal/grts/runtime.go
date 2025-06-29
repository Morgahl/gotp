package grts

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/application"
	"github.com/Morgahl/gotp/debug"
	"github.com/Morgahl/gotp/internal/ctx"
)

func Run(app application.Application) (err error) {
	ctx := ctx.Root()
	defer ctx.Cancel(fmt.Errorf("grts.Run: exiting"))

	wg := &sync.WaitGroup{}
	wg.Add(1)
	root := gotp.Spawn(_init(ctx, app, wg))
	root.Start()
	slog.DebugContext(ctx, "grts.Run: root process started", slog.Any("pid", root.ID()))

	<-ctx.Done()
	wg.Wait()

	slog.Info("grts.Run: exiting")
	return context.Cause(ctx)
}

func _init(rootCtx ctx.Cancellable, app application.Application, wg *sync.WaitGroup) gotp.RunFn {
	return func(p *gotp.Process) (reason error) {
		defer wg.Done()
		defer func() {
			reason = debug.Recover(recover(), "grts._init", reason)
			rootCtx.Cancel(reason)
		}()

		var root gotp.Supervisable
		var sup gotp.Supervised
		startUp := time.Now()
		if root, reason = app.Start(application.Normal()); reason != nil {
			slog.ErrorContext(rootCtx, "grts._init: failed to start application", slog.String("error", reason.Error()))
			return fmt.Errorf("failed to start application: %w", reason)
		} else if sup, reason = root.StartLink(p.ID()); reason != nil {
			slog.ErrorContext(rootCtx, "grts._init: failed to start supervision tree", slog.String("error", reason.Error()))
			return reason
		}
		appStart := time.Now()
		defer func(appStart time.Time) {
			slog.InfoContext(rootCtx, "grts._init: application exited", slog.Duration("after", time.Since(appStart)))
		}(appStart)
		slog.DebugContext(rootCtx, "grts._init: supervision tree started", slog.Any("pid", sup.ID()), slog.Duration("took", time.Since(startUp)))

		for {
			select {
			case <-rootCtx.Done():
				reason = context.Cause(rootCtx)
				goto EXIT
			case msg, ok := <-p.Receive():
				if !ok {
					slog.DebugContext(rootCtx, "grts._init: process channel closed")
					reason = fmt.Errorf("process channel closed unexpectedly")
					goto EXIT
				}
				slog.DebugContext(rootCtx, "grts._init: received message", slog.String("msg", fmt.Sprintf("%+v", msg)))
				switch msg := msg.(type) {
				case gotp.Exit:
					slog.DebugContext(rootCtx, "grts._init: received exit", slog.String("msg", fmt.Sprintf("%+v", msg)))
					reason = msg.Unwrap()
					goto EXIT
				default:
					slog.WarnContext(rootCtx, "grts._init: ignoring message", slog.String("msg", fmt.Sprintf("%+v", msg)))
				}
			}
		}

	EXIT:
		shutDown := time.Now()
		sup.Send(gotp.NewExit(sup.ID(), reason))
		slog.DebugContext(rootCtx, "grts._init: waiting for exit messages", slog.Duration("took", time.Since(shutDown)))
		for msg := range p.Receive() {
			switch msg := msg.(type) {
			case gotp.Exit:
				slog.DebugContext(rootCtx, "grts._init: received exit", slog.String("msg", fmt.Sprintf("%+v", msg)), slog.Duration("took", time.Since(shutDown)))
				return msg.Unwrap()
			default:
				slog.WarnContext(rootCtx, "grts._init: ignoring message", slog.String("msg", fmt.Sprintf("%+v", msg)), slog.Duration("took", time.Since(shutDown)))
			}
		}

		slog.InfoContext(rootCtx, "grts._init: exiting", slog.Duration("took", time.Since(shutDown)))
		return reason
	}
}
