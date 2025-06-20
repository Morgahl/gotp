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
	root := gotp.Spawn(initLoop(ctx, app, wg))
	slog.DebugContext(ctx, "grts.Run: root process started", slog.Any("pid", root.PID()))

	<-ctx.Done()
	wg.Wait()

	slog.Info("grts.Run: exiting")
	return context.Cause(ctx)
}

func initLoop(rootCtx ctx.Cancellable, app application.Application, wg *sync.WaitGroup) gotp.RunFn {
	return func(p *gotp.Process) (reason error) {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				reason = debug.Catch(reason, r)
			}
			rootCtx.Cancel(reason)
		}()

		var root gotp.Supervisable
		var sup gotp.Supervised
		startUp := time.Now()
		if root, reason = app.Start(application.Normal()); reason != nil {
			slog.ErrorContext(rootCtx, "grts.initLoop: failed to start application", slog.String("error", reason.Error()))
			return fmt.Errorf("failed to start application: %w", reason)
		} else if sup, reason = root.StartLink(p.PID()); reason != nil {
			slog.ErrorContext(rootCtx, "grts.initLoop: failed to start supervision tree", slog.String("error", reason.Error()))
			return reason
		}
		slog.DebugContext(rootCtx, "grts.initLoop: supervision tree started", slog.Any("pid", sup.PID()), slog.Duration("took", time.Since(startUp)))

		for {
			select {
			case <-rootCtx.Done():
				reason = context.Cause(rootCtx)
				goto EXIT
			case msg, ok := <-p.Receive():
				if !ok {
					slog.DebugContext(rootCtx, "grts.initLoop: process channel closed")
					reason = fmt.Errorf("process channel closed unexpectedly")
					goto EXIT
				}
				slog.DebugContext(rootCtx, "grts.initLoop: received message", slog.String("msg", fmt.Sprintf("%+v", msg)))
				switch msg := msg.(type) {
				case gotp.Exit:
					slog.DebugContext(rootCtx, "grts.initLoop: received exit", slog.String("msg", fmt.Sprintf("%+v", msg)))
					reason = msg.Unwrap()
					goto EXIT
				default:
					slog.WarnContext(rootCtx, "grts.initLoop: ignoring message", slog.String("msg", fmt.Sprintf("%+v", msg)))
				}
			}
		}

	EXIT:
		shutDown := time.Now()
		sup.Send(gotp.NewExit(sup.PID(), reason), 0)
		slog.DebugContext(rootCtx, "grts.initLoop: waiting for exit messages", slog.Duration("took", time.Since(shutDown)))
		for msg := range p.Receive() {
			switch msg := msg.(type) {
			case gotp.Exit:
				slog.DebugContext(rootCtx, "grts.initLoop: received exit", slog.String("msg", fmt.Sprintf("%+v", msg)), slog.Duration("took", time.Since(shutDown)))
				return msg.Unwrap()
			default:
				slog.WarnContext(rootCtx, "grts.initLoop: ignoring message", slog.String("msg", fmt.Sprintf("%+v", msg)), slog.Duration("took", time.Since(shutDown)))
			}
		}

		slog.InfoContext(rootCtx, "grts.initLoop: exiting", slog.Duration("took", time.Since(shutDown)))
		return reason
	}
}
