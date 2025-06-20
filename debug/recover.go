package debug

import (
	"fmt"
)

type Caught struct {
	err error
	r   error
}

func Catch(err error, r any) Caught {
	switch v := r.(type) {
	case Thrown:
		return Caught{err: err, r: fmt.Errorf("\n%s\n%s", v.Message, v.Stack)}
	case error:
		return Caught{err: err, r: fmt.Errorf("panic: %w", v)}
	default:
		return Caught{err: err, r: fmt.Errorf("panic: %v", v)}
	}
}

func (r Caught) Error() string {
	if r.err == nil {
		return fmt.Sprintf("Recovered{%v}", r.r)
	}
	return fmt.Sprintf("Recovered{%v, %v}", r.err, r.r)
}
