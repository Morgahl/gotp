package main

import (
	"context"
	"log"

	"github.com/Morgahl/gotp/application"
	"github.com/Morgahl/gotp/examples/foo"
	"github.com/Morgahl/gotp/internal/ctx"
)

func main() {
	ctx := ctx.NewRoot(context.Background())
	_, err := foo.Start(application.Normal()).StartLink(ctx, 0)
	if err != nil {
		log.Fatalf("failed to start application: %v", err)
	}

	<-ctx.Done()
	log.Printf("shutting down: %v", ctx.Err())
}
