package ctx

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

type Cancellable interface {
	context.Context
	Cancel(error)
}

var _ context.Context = &Context{}

type Context struct {
	context.Context
	cancel context.CancelCauseFunc
}

func Root(ctx context.Context) *Context {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancelCause(ctx)
	go func(ctx context.Context, cancel context.CancelCauseFunc) {
		select {
		case <-ctx.Done():
			slog.DebugContext(ctx, "root context done", "reason", context.Cause(ctx))
			return
		case s := <-sig:
			slog.InfoContext(ctx, "signal received", "signal", s)
			cancel(fmt.Errorf("signal: %s", s))
		}
	}(ctx, cancel)
	return &Context{ctx, cancel}
}

func (r *Context) Cancel(err error) { r.cancel(err) }
