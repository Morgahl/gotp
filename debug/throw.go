package debug

import (
	"fmt"
	"log/slog"
	"runtime"
)

type Thrown struct {
	Message string
	Stack   []string
}

func (p Thrown) Error() string {
	return fmt.Sprintf("%s\n\n%s", p.Message, p.Stack)
}

func (p Thrown) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("message", p.Message),
		slog.Any("stack", p.Stack))
}

//go:noreturn
func Throw(message string) Thrown {
	return Thrown{
		Message: message,
		Stack:   callerStack(3)}
}

//go:noreturn
func ThrowF(format string, args ...any) Thrown {
	return Thrown{
		Message: fmt.Sprintf(format, args...),
		Stack:   callerStack(3)}
}

func callerStack(skip int) []string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(skip, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	var b = make([]string, 0, 10)
	for {
		frame, more := frames.Next()
		b = append(b, fmt.Sprintf("%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line))
		if !more {
			break
		}
	}
	return b[:len(b):len(b)]
}
