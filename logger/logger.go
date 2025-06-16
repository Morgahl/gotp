package logger

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

const (
	TIME_FORMAT = "2006-01-02 15:04:05.000000"

	DEBUG_PREFIX = "\x1b[36m" // Cyan
	INFO_PREFIX  = "\x1b[32m" // Green
	WARN_PREFIX  = "\x1b[33m" // Yellow
	ERROR_PREFIX = "\x1b[31m" // Red
	RESET_PREFIX = "\x1b[0m"  // Reset color
)

type textHandler struct {
	mu      sync.Mutex
	w       io.Writer
	opts    slog.HandlerOptions
	timeBuf *bytes.Buffer
}

func TextHandler(w io.Writer, opts *slog.HandlerOptions) slog.Handler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &textHandler{
		w:       w,
		opts:    *opts,
		timeBuf: &bytes.Buffer{},
	}
}

func (h *textHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.opts.Level.Level()
}

func (h *textHandler) Handle(_ context.Context, r slog.Record) error {
	var buf bytes.Buffer

	buf.WriteString(colorLevelPrefix(r.Level))
	buf.WriteString(r.Time.Format(TIME_FORMAT))
	buf.WriteByte(' ')

	buf.WriteString(r.Level.String())
	buf.WriteByte(' ')

	if h.opts.AddSource && r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		buf.WriteString(trimProjectPath(f.File))
		buf.WriteByte(':')
		buf.WriteString(strconv.Itoa(f.Line))
		buf.WriteByte(' ')
	}

	if r.Message != "" {
		buf.WriteString(r.Message)
	}

	appendAttr := func(a slog.Attr) {
		if !a.Equal(slog.Attr{}) {
			buf.WriteByte(' ')
			writeAttr(&buf, "", a)
		}
	}

	r.Attrs(func(a slog.Attr) bool {
		if h.opts.ReplaceAttr != nil {
			a = h.opts.ReplaceAttr(nil, a)
		}
		appendAttr(a)
		return true
	})

	buf.WriteString(RESET_PREFIX)

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.w.Write(buf.Bytes())
	if err == nil {
		_, err = h.w.Write([]byte{'\n'})
	}
	return err
}

func (h *textHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// newOpts := h.opts
	// newOpts.AddSource = false
	// newHandler := &textHandler{
	// 	w:       h.w,
	// 	opts:    newOpts,
	// 	timeBuf: h.timeBuf,
	// }
	// return newHandler
	panic("WithAttrs not implemented for textHandler")
}

func (h *textHandler) WithGroup(name string) slog.Handler {
	// newHandler := *h
	// return &newHandler
	panic("WithGroup not implemented for textHandler")
}

func writeAttr(buf *bytes.Buffer, prefix string, a slog.Attr) {
	key := a.Key
	if prefix != "" {
		key = prefix + "." + key
	}

	val := a.Value
	switch val.Kind() {
	case slog.KindGroup:
		for _, ga := range val.Group() {
			writeAttr(buf, key, ga)
		}
	case slog.KindString:
		buf.WriteString(key)
		buf.WriteByte('=')
		buf.WriteString(strconv.Quote(val.String()))
	default:
		buf.WriteString(key)
		buf.WriteByte('=')
		buf.WriteString(fmt.Sprint(val.Any()))
	}
}

var (
	projectRootDir string
	initRootOnce   sync.Once
)

func trimProjectPath(path string) string {
	initRootOnce.Do(func() {
		_, file, _, ok := runtime.Caller(0)
		if !ok {
			return
		}
		// Find unique suffix of this file to isolate the root prefix
		const marker = "logger/logger.go" // adjust if filename is different
		if idx := strings.LastIndex(file, marker); idx != -1 {
			projectRootDir = file[:idx]
		}
	})

	if projectRootDir != "" && strings.HasPrefix(path, projectRootDir) {
		return path[len(projectRootDir):]
	}
	return path
}

func colorLevelPrefix(level slog.Level) string {
	switch level {
	case slog.LevelDebug:
		return DEBUG_PREFIX
	case slog.LevelInfo:
		return INFO_PREFIX
	case slog.LevelWarn:
		return WARN_PREFIX
	case slog.LevelError:
		return ERROR_PREFIX
	default:
		return RESET_PREFIX
	}
}
