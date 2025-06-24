package debug

import (
	"bytes"
	"fmt"
	"runtime/debug"
	"strings"
)

type Caught struct {
	err   error
	r     error
	stack []string
}

func Catch(err error, r any) Caught {
	switch v := r.(type) {
	case Thrown:
		return Caught{err: err, r: fmt.Errorf("\n%s\n%s", v.Message, v.Stack), stack: v.Stack}
	case error:
		return Caught{err: err, r: fmt.Errorf("panic: %w", v), stack: f(debug.Stack())}
	default:
		return Caught{err: err, r: fmt.Errorf("panic: %v", v), stack: f(debug.Stack())}
	}
}

func (r Caught) Error() string {
	if r.err == nil {
		return fmt.Sprintf("Caught{%v, %v}", r.r, p(r.stack))
	}
	return fmt.Sprintf("Caught{%v, %v, %v}", r.err, r.r, p(r.stack))
}

func f(stack []byte) []string {
	lines := make([]string, 0, 10)
	for _, line := range bytes.Split(stack, []byte{'\n'}) {
		if len(line) > 0 {
			lines = append(lines, string(line))
		}
	}
	return lines
}

func p(stack []string) string {
	if len(stack) == 0 {
		return ""
	}
	var b bytes.Buffer
	for _, line := range stack {
		b.WriteByte(' ')
		b.WriteString(strings.TrimPrefix(line, "\t"))
		b.WriteByte('\n')
	}
	return b.String()
}
