package main

import (
	"fmt"
	"log/slog"

	"github.com/Morgahl/gotp/logger"
	"github.com/Morgahl/gotp/old/internal/grts"

	"github.com/Morgahl/gotp/old/examples/game"
)

func init() {
	logger.ConfigFromEnv()
}

func main() {
	slog.Info("main: starting application",
		slog.Any("name", game.Game{}.Name()),
		slog.Any("version", game.Game{}.Version()))
	reason := grts.Run(game.Game{})
	slog.Info("main: application exited", slog.Any("reason", reason))
	fmt.Println(reason)
}
