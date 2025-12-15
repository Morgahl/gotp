package dbg

import (
	"fmt"
	"log/slog"
	"runtime"
	"strings"
)

var (
	thrownTypeString = fmt.Sprintf("%T", Thrown{})
)

type Thrown struct {
	Message string
	Stack   string
}

func (p Thrown) Error() string {
	return fmt.Sprintf("%s\n\n%s", p.Message, p.Stack)
}

func (p Thrown) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("type", thrownTypeString),
		slog.String("message", p.Message),
		slog.String("stack", p.Stack))
}

func Throw(format string, args ...any) {
	throw(1, format, args...)
}

func throw(skip int, format string, args ...any) {
	panic(Thrown{
		Message: fmt.Sprintf(format, args...),
		Stack:   CallerStack(skip + 3)})
}

func CallerStack(skip int) string {
	const depth = 36
	var pcs [depth]uintptr
	n := runtime.Callers(skip, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	var b strings.Builder
	b.Grow(64 * depth)
	for frame, more := frames.Next(); more; frame, more = frames.Next() {
		fmt.Fprintf(&b, "%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line)
	}
	return b.String()
}
