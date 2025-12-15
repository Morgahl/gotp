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
	server := node.NewServer(rpc.SERVER_NAME, rpc.COOKIE)
	go server.Start()
	time.Sleep(time.Second * 5)
	err := server.Connect(rpc.CLIENT_NAME, rpc.COOKIE, true)
	if err != nil {
		slog.Error("Failed to connect to other node", "err", err)
		return
	}
	visible, err := server.List("visible")
	if err != nil {
		slog.Error("Failed to list visible nodes", "err", err)
		return
	}
	slog.Info("Visible nodes", "nodes", visible)
	hidden, err := server.List("hidden")
	if err != nil {
		slog.Error("Failed to list hidden nodes", "err", err)
		return
	}
	slog.Info("Hidden nodes", "nodes", hidden)
	<-ctx.Done()
}
