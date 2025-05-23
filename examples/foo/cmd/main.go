package main

import (
	"context"
	"log"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/application"
	"github.com/Morgahl/gotp/examples/foo"
	"github.com/Morgahl/gotp/internal/ctx"
)

func main() {
	ctx := ctx.NewRoot(context.Background())
	// TODO: WE NEED TO SET THIS UP TO SEND MESSAGES DOWN THE SUPERVISION TREE TO MANAGE GRACEFULLY ROLLING IT UP
	supTree, err := foo.Start(application.Normal()).StartLink(context.Background(), gotp.RootPID())
	if err != nil {
		log.Fatalf("failed to start application: %v", err)
	}

	<-ctx.Done()
	log.Printf("shutting down: %v", ctx.Err())

	gotp.Send(context.Background(), supTree.ID(), gotp.NewExit(supTree.ID(), ctx.Err()))

	// TODO: something better then a wait ... let's somehow check the state of the main process
	time.Sleep(time.Second)
	log.Printf("shut down complete")
}
