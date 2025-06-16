package grts

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/application"
	"github.com/Morgahl/gotp/internal/ctx"
)

const (
	FORCE_EXIT_AFTER = 30 * time.Second
)

func Setup() {
	switch os.Getenv("LOG_LEVEL") {
	case "debug", "DEBUG":
		slog.SetLogLoggerLevel(slog.LevelDebug)
	case "info", "INFO":
		slog.SetLogLoggerLevel(slog.LevelInfo)
	case "warn", "WARN":
		slog.SetLogLoggerLevel(slog.LevelWarn)
	case "error", "ERROR":
		slog.SetLogLoggerLevel(slog.LevelError)
	default:
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}
}

func Run(app application.Application) (err error) {
	ctx := ctx.Root()
	defer ctx.Cancel(fmt.Errorf("main: exiting"))

	wg := &sync.WaitGroup{}
	wg.Add(1)
	root := gotp.Spawn(initLoop(ctx, app, wg))
	slog.InfoContext(ctx, "root process started", "pid", root.PID())

	<-ctx.Done()
	wg.Wait()

	slog.Info("main: exiting")
	return context.Cause(ctx)
}

func initLoop(rootCtx ctx.Cancellable, app application.Application, wg *sync.WaitGroup) gotp.RunFn {
	return func(p *gotp.Process) (reason error) {
		defer wg.Done()
		defer func() {
			gotp.Recover(&reason)
			slog.InfoContext(rootCtx, "initLoop: exiting", "reason", reason)
			// Handle shutdown of main loop
			rootCtx.Cancel(reason)
		}()

		var root gotp.Supervisable
		var sup gotp.Supervised
		if root, reason = app.Start(application.Normal()); reason != nil {
			slog.ErrorContext(rootCtx, "initLoop: failed to start application", "error", reason)
			return fmt.Errorf("failed to start application: %w", reason)
		} else if sup, reason = root.StartLink(p.PID(), 0); reason != nil {
			slog.ErrorContext(rootCtx, "initLoop: failed to start supervision tree", "error", reason)
			return reason
		}
		slog.InfoContext(rootCtx, "initLoop: supervision tree started", "id", sup.PID())

		for {
			select {
			case <-rootCtx.Done():
				goto EXIT
			case msg, ok := <-p.Receive():
				if !ok {
					slog.InfoContext(rootCtx, "initLoop: process channel closed")
					goto EXIT
				}
				slog.InfoContext(rootCtx, "initLoop: received message", "msg", msg)
				switch msg := msg.(type) {
				case gotp.Exit:
					slog.InfoContext(rootCtx, "initLoop: received exit", "msg", msg)
					reason = msg.Unwrap()
					goto EXIT
				default:
					slog.WarnContext(rootCtx, "initLoop: ignoring message", "msg", msg)
				}
			}
		}

	EXIT:
		sup.Send(gotp.NewExit(sup.PID(), reason), 0)
		slog.InfoContext(rootCtx, "initLoop: waiting for messages")
		for msg := range p.Receive() {
			slog.InfoContext(rootCtx, "initLoop: received message", "msg", msg)
			switch msg := msg.(type) {
			case gotp.Exit:
				slog.InfoContext(rootCtx, "initLoop: received exit", "msg", msg)
				return msg.Unwrap()
			default:
				slog.WarnContext(rootCtx, "initLoop: ignoring message", "msg", msg)
			}
		}

		slog.InfoContext(rootCtx, "initLoop: exiting")
		return reason
	}
}
