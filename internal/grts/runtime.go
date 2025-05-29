package grts

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/application"
	"github.com/Morgahl/gotp/internal/ctx"
)

const (
	// TODO: make this at least 30
	FORCE_EXIT_AFTER = 1 * time.Second
)

func Init() {
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

func Main(app application.Application) (err error) {
	ctx := ctx.Root()
	defer ctx.Cancel(fmt.Errorf("main: exiting"))

	var sup gotp.Supervisable
	if sup, err = app.Start(application.Normal()); err != nil {
		slog.ErrorContext(ctx, "main: failed to start application", "error", err)
		return fmt.Errorf("failed to start application: %w", err)
	}

	root := gotp.Spawn(mainLoop(ctx, sup), 0)
	slog.InfoContext(ctx, "root process started", "pid", root.PID())

	<-ctx.Done()

	// Handle shutdown gracefully

	forceExitAfter := time.After(FORCE_EXIT_AFTER)
	checkExitTicker := time.NewTicker(100 * time.Millisecond)
	defer checkExitTicker.Stop()

EXIT:
	for {
		select {
		case <-forceExitAfter:
			slog.Info("force exit after", "duration", FORCE_EXIT_AFTER)
			break EXIT

		case <-checkExitTicker.C:
			slog.Info("checking if root exited")
			if root.Exited() {
				slog.Info("root exited", "reason", root.Reason())
				break EXIT
			}
		}
	}
	slog.Info("main: exiting")
	return context.Cause(ctx)
}

func mainLoop(rootCtx ctx.Cancellable, root gotp.Supervisable) gotp.RunFn {
	return func(p *gotp.Process, in <-chan gotp.Msg) (reason error) {
		defer func() {
			if r := recover(); r != nil {
				slog.ErrorContext(rootCtx, "mainLoop: panic", "reason", r)
				reason = fmt.Errorf("panic: %v", r)
			}
			slog.InfoContext(rootCtx, "mainLoop: exiting", "reason", reason)
			// Handle shutdown of main loop
			rootCtx.Cancel(reason)
		}()

		slog.Info("mainLoop: starting supervision tree")
		var sup gotp.Supervisable
		if sup, reason = root.StartLink(p.PID(), 0); reason != nil {
			slog.Error("mainLoop: failed to start supervision tree", "error", reason)
			return reason
		}
		slog.Info("mainLoop: supervision tree started", "id", sup.PID())

		<-rootCtx.Done()
		reason = context.Cause(rootCtx)
		slog.Info("shutting down", "reason", reason)

		sup.Send(gotp.NewExit(root.PID(), reason), 0)
		slog.Info("mainLoop: waiting for messages")
		for msg := range in {
			slog.Info("mainLoop: received message", "msg", msg)
			switch msg := msg.(type) {
			case gotp.Exit:
				slog.Info("mainLoop: received exit", "msg", msg)
				return msg.Unwrap()
			default:
				slog.Warn("mainLoop: ignoring message", "msg", msg)
			}
		}

		slog.Info("mainLoop: exiting")
		return reason
	}
}
