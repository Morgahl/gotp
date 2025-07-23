package dbg

import (
	"errors"
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

// Recover expects the first arguement to be the result of calling `recover()`. If the result is nil, it returns the
// provided error. If the result is not nil, it constructs a [Recovered] error containing the panic information merged
// with the provided message, error, and the stack trace of the point of recovery.
func Recover(r any, message string, err error) error {
	if r == nil {
		return err
	}
	rec := Recovered{message: message, err: err, stack: debug.Stack()}
	switch v := r.(type) {
	case Thrown:
		rec.r = v
	case error:
		rec.r = fmt.Errorf("%s: panic error: %w", message, v)
	default:
		rec.r = fmt.Errorf("%s: panic unknown: %v", message, v)
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

func (r Recovered) Is(err error) bool {
	return errors.Is(r.err, err) || errors.Is(r.r, err)
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
