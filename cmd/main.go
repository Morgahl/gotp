package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Morgahl/gotp/application"
	"github.com/Morgahl/gotp/examples/foo"
)

func main() {
	ctx := rootCtx(context.Background())
	_, err := foo.Start(application.Normal()).StartLink(ctx, 0)
	if err != nil {
		log.Fatalf("failed to start application: %v", err)
	}

	<-ctx.Done()
	log.Printf("shutting down: %v", ctx.Err())
}

type RootContext struct {
	context.Context
	context.CancelFunc
	os.Signal
}

func (r *RootContext) Err() error {
	if r.Signal != nil {
		return fmt.Errorf("exit with signal: %v", r.Signal)
	}
	return r.Context.Err()
}

var _ context.Context = &RootContext{}

func rootCtx(ctx context.Context) *RootContext {
	ctx, cancel := context.WithCancel(ctx)
	root := &RootContext{
		Context:    ctx,
		CancelFunc: cancel,
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-ctx.Done():
		case s := <-sig:
			log.Printf("received signal: %v", s)
			cancel()
			root.Signal = s
		}
	}()
	return root
}
