package main

import (
	"fmt"
	"log/slog"

	"github.com/Morgahl/gotp/internal/grts"
	"github.com/Morgahl/gotp/logger"

	"github.com/Morgahl/gotp/examples/foo"
)

func init() {
	logger.ConfigFromEnv()
}

func main() {
	slog.Info("main: starting application",
		slog.Any("name", foo.WarGamesApplication{}.Name()),
		slog.Any("version", foo.WarGamesApplication{}.Version()))
	reason := grts.Run(foo.WarGamesApplication{})
	slog.Info("main: application exited", slog.Any("reason", reason))
	fmt.Println(reason)
}
