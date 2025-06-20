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

func Root(vals ...any) *Context {
	if len(vals)%2 != 0 {
		panic("Root: expected even number of values for key-value pairs")
	}
	ctx := context.Background()
	for i := 0; i < len(vals); i += 2 {
		ctx = context.WithValue(ctx, vals[i], vals[i+1])
	}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancelCause(ctx)
	go func(ctx context.Context, cancel context.CancelCauseFunc) {
		select {
		case <-ctx.Done():
			slog.DebugContext(ctx, "root context done", slog.String("reason", context.Cause(ctx).Error()))
			return
		case s := <-sig:
			slog.InfoContext(ctx, "signal received, shutting down", slog.Any("signal", s))
			cancel(newShutdown(s))
		}
	}(ctx, cancel)
	return &Context{ctx, cancel}
}

func (r *Context) Cancel(err error) { r.cancel(err) }

type Shutdown struct {
	reason os.Signal
}

func newShutdown(signal os.Signal) Shutdown {
	return Shutdown{signal}
}

func (s Shutdown) Error() string {
	return fmt.Sprintf("Shutdown{reason: %v}", s.reason)
}
