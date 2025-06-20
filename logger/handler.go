package logger

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"strconv"
	"sync"
)

const (
	TIME_FORMAT = "2006-01-02 15:04:05.000000"

	DEBUG_PREFIX = "\x1b[36m" // Cyan
	INFO_PREFIX  = "\x1b[32m" // Green
	WARN_PREFIX  = "\x1b[33m" // Yellow
	ERROR_PREFIX = "\x1b[31m" // Red
	FILE_PREFIX  = "\x1b[90m" // Gray
	RESET_PREFIX = "\x1b[0m"  // Reset color
)

type textHandler struct {
	mu         sync.Mutex
	w          io.Writer
	opts       slog.HandlerOptions
	colors     bool
	scratch    []byte
	messageBuf bytes.Buffer
}

func TextHandler(w io.Writer, colors bool, opts *slog.HandlerOptions) slog.Handler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &textHandler{
		w:          w,
		opts:       *opts,
		colors:     colors,
		scratch:    make([]byte, 0, 32),
		messageBuf: bytes.Buffer{},
	}
}

func (h *textHandler) Enabled(_ context.Context, level slog.Level) bool {
	if h == nil {
		panic("textHandler is nil")
	} else if h.w == nil {
		panic("textHandler has no writer")
	} else if h.opts.Level == nil {
		panic("textHandler has no level set")
	}
	return level >= h.opts.Level.Level()
}

func (h *textHandler) Handle(_ context.Context, r slog.Record) error {
	h.messageBuf.Reset()
	applyColor(&h.messageBuf, r.Level, h.colors)
	h.scratch = r.Time.AppendFormat(h.scratch[:0], TIME_FORMAT)
	h.messageBuf.Write(h.scratch)
	h.messageBuf.WriteByte(' ')
	h.messageBuf.WriteString(r.Level.String())
	h.messageBuf.WriteByte(' ')
	h.messageBuf.WriteString(r.Message)
	r.Attrs(func(a slog.Attr) bool {
		if h.opts.ReplaceAttr != nil {
			a = h.opts.ReplaceAttr(nil, a)
		}
		appendAttr(&h.messageBuf, a)
		return true
	})

	if h.opts.AddSource && r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		h.messageBuf.WriteString(FILE_PREFIX)
		h.messageBuf.WriteByte(' ')
		h.messageBuf.WriteString(f.File)
		h.messageBuf.WriteByte(':')
		h.messageBuf.WriteString(strconv.Itoa(f.Line))
	}

	resetColor(&h.messageBuf, h.colors)

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.w.Write(h.messageBuf.Bytes())
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

func appendAttr(buf *bytes.Buffer, a slog.Attr) {
	if !a.Equal(slog.Attr{}) {
		buf.WriteByte(' ')
		writeAttr(buf, "", a)
	}
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
			writeAttr(buf, prefix+key, ga)
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

func applyColor(buf *bytes.Buffer, level slog.Level, shouldColor bool) (int, error) {
	if !shouldColor {
		return 0, nil
	}
	switch level {
	case slog.LevelDebug:
		return buf.WriteString(DEBUG_PREFIX)
	case slog.LevelInfo:
		return buf.WriteString(INFO_PREFIX)
	case slog.LevelWarn:
		return buf.WriteString(WARN_PREFIX)
	case slog.LevelError:
		return buf.WriteString(ERROR_PREFIX)
	}
	return 0, nil
}

func resetColor(buf *bytes.Buffer, shouldColor bool) (int, error) {
	if !shouldColor {
		return 0, nil
	}
	return buf.WriteString(RESET_PREFIX)
}
