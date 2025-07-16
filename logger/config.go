package logger

import (
	"context"
	"log/slog"
	"os"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/tint"
)

const (
	TIME_FORMAT = "2006-01-02 15:04:05.000000"
)

func ConfigFromEnv(args ...any) {
	stdout := os.Stdout
	opts := tint.Options{
		AddSource:  true,
		Level:      slog.LevelInfo,
		TimeFormat: TIME_FORMAT,
	}

	switch os.Getenv("LOG_SOURCE") {
	case "false", "FALSE", "0":
		opts.AddSource = false
	}

	switch os.Getenv("LOG_LEVEL") {
	case "debug", "DEBUG":
		opts.Level = slog.LevelDebug
	case "info", "INFO":
		opts.Level = slog.LevelInfo
	case "warn", "WARN":
		opts.Level = slog.LevelWarn
	case "error", "ERROR":
		opts.Level = slog.LevelError
	}

	colors := true
	switch os.Getenv("NO_COLOR") {
	case "true", "TRUE", "1":
		colors = false
	}
	opts.NoColor = !colors

	handler := tint.NewHandler(stdout, &opts)
	// handler := slog.NewJSONHandler(stdout, &slog.HandlerOptions{
	// 	AddSource: opts.AddSource,
	// 	Level:     opts.Level,
	// })
	handler = NewGOTPContextHandler(handler, "pid", "name")
	logger := slog.New(handler).With(args...)
	slog.SetDefault(logger)
}

type GOTPContextHandler struct {
	keys []gotp.Atom
	slog.Handler
}

func NewGOTPContextHandler(handler slog.Handler, keys ...gotp.Atom) *GOTPContextHandler {
	return &GOTPContextHandler{
		keys:    keys,
		Handler: handler,
	}
}

func (h *GOTPContextHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, key := range h.keys {
		if v := ctx.Value(key); v != nil {
			r.AddAttrs(slog.Any(key.String(), v))
		}
	}
	return h.Handler.Handle(ctx, r)
}
