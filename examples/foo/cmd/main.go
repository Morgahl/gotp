package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/application"
	"github.com/Morgahl/gotp/examples/foo"
	"github.com/Morgahl/gotp/supervisor"
)

const (
	FORCE_EXIT_AFTER = 30 * time.Second
)

func init() {
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

func main() {
	ctx := newRoot()
	defer ctx.Cancel(fmt.Errorf("main: exiting"))

	root := gotp.Spawn(mainLoop(ctx, foo.Start(application.Normal())), 0)
	slog.InfoContext(ctx, "root process started", "pid", root.ID())

	<-ctx.Done()

	forceExitAfter := time.After(FORCE_EXIT_AFTER)
	checkExitTicker := time.NewTicker(100 * time.Millisecond)
	defer checkExitTicker.Stop()

	for {
		select {
		case <-forceExitAfter:
			slog.Info("force exit after", "duration", FORCE_EXIT_AFTER)
			return

		case <-checkExitTicker.C:
			slog.Info("checking if root exited")
			if root.Exited() {
				slog.Info("root exited", "reason", root.Reason())
				return
			}
		}
	}
}

type Shutdown struct {
	reason os.Signal
}

func newShutdown(signal os.Signal) Shutdown {
	return Shutdown{signal}
}

func (s Shutdown) Error() string {
	return fmt.Sprintf("Shutdown{reason: %v}", s.reason)
}

func mainLoop(rootCtx context.Context, root supervisor.Supervisor) gotp.RunFn {
	return func(p *gotp.Process, in <-chan gotp.Msg) (reason error) {
		defer func() {
			if r := recover(); r != nil {
				slog.ErrorContext(rootCtx, "mainLoop: panic", "reason", r)
				reason = fmt.Errorf("panic: %v", r)
			}
			slog.InfoContext(rootCtx, "mainLoop: exiting", "reason", reason)
		}()

		slog.Info("mainLoop: starting supervision tree")
		var sup supervisor.Supervisable
		if sup, reason = root.StartLink(p.ID(), 0); reason != nil {
			slog.Error("mainLoop: failed to start supervision tree", "error", reason)
			return reason
		}
		slog.Info("mainLoop: supervision tree started", "id", sup.ID())

		<-rootCtx.Done()
		reason = context.Cause(rootCtx)
		slog.Info("shutting down", "reason", reason)

		sup.Send(gotp.NewExit(root.ID(), reason), 0)
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

var _ context.Context = &iCtx{}

type iCtx struct {
	context.Context
	cancel context.CancelCauseFunc
}

func newRoot() *iCtx {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancelCause(context.Background())
	go func(ctx context.Context, cancel context.CancelCauseFunc) {
		select {
		case <-ctx.Done():
			slog.Info("root context done", "reason", context.Cause(ctx))
			return
		case s := <-sig:
			slog.Info("signal received", "signal", s)
			cancel(newShutdown(s))
		}
	}(ctx, cancel)
	return &iCtx{ctx, cancel}
}

func (r *iCtx) Cancel(err error) { r.cancel(err) }
