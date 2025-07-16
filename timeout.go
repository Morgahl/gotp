package gotp

import (
	"fmt"
	"time"
)

type Timeout struct {
	reason error
}

func NewTimeout(dur time.Duration) Timeout {
	return Timeout{reason: fmt.Errorf("timeout after %s", dur)}
}

func (t Timeout) Error() string {
	return fmt.Sprintf("Timeout{reason: %s}", t.reason)
}
