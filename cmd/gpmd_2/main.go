package main

import (
	"github.com/Morgahl/gotp/internal/ctx"
	gpmd "github.com/Morgahl/gotp/internal/gpmd_2"
	"github.com/Morgahl/gotp/logger"
)

func init() {
	logger.ConfigFromEnv()
}

func main() {
	ctx := ctx.Root()
	server := gpmd.NewServer(ctx)
	go server.Start()
	<-ctx.Done()
	server.Stop()
}
