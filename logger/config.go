package logger

import (
	"log/slog"
	"os"
)

var rootLogger *slog.Logger

func ConfigFromEnv(args ...any) {
	opts := slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
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

	handler := TextHandler(os.Stdout, colors, &opts)
	rootLogger = slog.New(handler).With(args...)
}

func SetGlobalDefaultLogger() { slog.SetDefault(rootLogger) }
