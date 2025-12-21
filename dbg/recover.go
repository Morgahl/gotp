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
	stack   []byte
}

const thrownFmt = `
%s

recovery message:
%s

recovery stack:
%s
`

func (r Recovered) String() string {
	if thrown, ok := r.err.(Thrown); ok {
		return fmt.Sprintf(thrownFmt, thrown, r.message, string(r.stack))
	}

	var b strings.Builder
	if r.message != "" {
		fmt.Fprintf(&b, "%s\n", r.message)
	}
	fmt.Fprintf(&b, "%v\n\n", r.err)
	b.Write(r.stack)
	return b.String()
}

// Recover expects the first arguement to be the result of calling `recover()`. If the result is nil, it returns the
// provided error. If the result is not nil, it constructs a [Recovered] error containing the panic information merged
// with the provided message, error, and the stack trace of the point of recovery.
func Recover(r any, message string, err error) error {
	if r == nil {
		return err
	}
	rec := Recovered{message: message, stack: debug.Stack()}
	switch v := r.(type) {
	case Thrown:
		rec.err = v
	case error:
		rec.err = fmt.Errorf("%s: panic error: %w", message, v)
	default:
		rec.err = fmt.Errorf("%s: panic unknown: %v", message, v)
	}
	return rec
}

func (r Recovered) Error() string {
	return r.String()
}

func (r Recovered) Is(err error) bool {
	return errors.Is(r.err, err)
}

func (r Recovered) LogValue() slog.Value {
	attrs := make([]slog.Attr, 0, 3)
	attrs = append(attrs, slog.String("type", recoveredTypeString))
	if r.err != nil {
		attrs = append(attrs, slog.Any("err", r.err))
	}
	attrs = append(attrs, slog.Any("recover", r.err))
	return slog.GroupValue(attrs...)
}
