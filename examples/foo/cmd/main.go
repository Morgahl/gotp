package main

import (
	"log/slog"

	"github.com/Morgahl/gotp/internal/grts"

	"github.com/Morgahl/gotp/examples/foo"
)

func init() {
	grts.Init()
}

func main() {
	slog.Info("main: starting application",
		"name", foo.FooApplication{}.Name(),
		"version", foo.FooApplication{}.Version())
	slog.Info("main: application exited", "reason", grts.Main(foo.FooApplication{}))
}
