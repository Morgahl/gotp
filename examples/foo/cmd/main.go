package main

import (
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
		// "name", foo.FooApplication{}.Name(),
		// "version", foo.FooApplication{}.Version())
		slog.Any("name", foo.FooApplication{}.Name()),
		slog.Any("version", foo.FooApplication{}.Version()),
	)
	reason := grts.Run(foo.FooApplication{})
	slog.Info("main: application exited", slog.Any("reason", reason))
}
