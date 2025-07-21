package main

import (
	"log/slog"
	"time"

	"github.com/Morgahl/gotp/internal/ctx"
	"github.com/Morgahl/gotp/internal/node"
	"github.com/Morgahl/gotp/logger"

	"github.com/Morgahl/gotp/cmd/rpc"
)

func init() {
	logger.ConfigFromEnv()
}

func main() {
	ctx := ctx.Root()
	server := node.NewServer(rpc.CLIENT_NAME, rpc.COOKIE)
	go server.Start()
	time.Sleep(time.Second * 5)
	err := server.Connect(rpc.SERVER_NAME, rpc.COOKIE, true)
	if err != nil {
		slog.Error("Failed to connect to other node", "err", err)
		return
	}
	visible := server.ListVisible()
	slog.Info("Visible nodes", "nodes", visible)
	hidden := server.ListHidden()
	slog.Info("Hidden nodes", "nodes", hidden)
	<-ctx.Done()
}
