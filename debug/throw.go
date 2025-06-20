package debug

import (
	"fmt"
	"runtime"
)

type Thrown struct {
	Message string
	Stack   []byte
}

func (p Thrown) Error() string {
	return fmt.Sprintf("%s\n\n%s", p.Message, p.Stack)
}

//go:noreturn
func Throw(message string) Thrown {
	return Thrown{
		Message: message,
		Stack:   callerStack(3), // skip Panic and runtime.Callers
	}
}

//go:noreturn
func ThrowF(format string, args ...any) Thrown {
	return Thrown{
		Message: fmt.Sprintf(format, args...),
		Stack:   callerStack(3), // skip Panicf and runtime.Callers
	}
}

func callerStack(skip int) []byte {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(skip, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	var b []byte
	for {
		frame, more := frames.Next()
		b = append(b, fmt.Sprintf("%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line)...)
		if !more {
			break
		}
	}
	return b
}
