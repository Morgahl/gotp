package debug

import (
	"fmt"
	"log/slog"
	"runtime/debug"
	"strings"
)

var (
	recoveredTypeString = fmt.Sprintf("%T", Recovered{})
)

type Recovered struct {
	message string
	err     error
	r       error
	stack   []byte
}

func Drop(r any) {}

func Recover(r any, message string, err error) error {
	if r == nil {
		return err
	}
	rec := Recovered{message: message, err: err}
	switch v := r.(type) {
	case Thrown:
		rec.r = v
	case error:
		rec.r = fmt.Errorf("%s: panic error: %w", message, v)
		rec.stack = debug.Stack()
	default:
		rec.r = fmt.Errorf("%s: panic unknown: %v", message, v)
		rec.stack = debug.Stack()
	}
	return rec
}

func (r Recovered) Error() string {
	var b strings.Builder
	if r.err == nil {
		fmt.Fprintf(&b, "Recovered{%v", r.r)
	} else {
		fmt.Fprintf(&b, "Recovered{%v, %v", r.err, r.r)
	}
	b.WriteByte('}')
	return b.String()
}

func (r Recovered) LogValue() slog.Value {
	attrs := make([]slog.Attr, 0, 3)
	attrs = append(attrs, slog.String("type", recoveredTypeString))
	if r.err != nil {
		attrs = append(attrs, slog.Any("err", r.err))
	}
	attrs = append(attrs, slog.Any("recover", r.r))
	return slog.GroupValue(attrs...)
}
