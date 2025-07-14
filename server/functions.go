package server

import (
	"time"

	"github.com/Morgahl/gotp/process"
)

const (
	DEFAULT_TIMEOUT = 5 * time.Second
)

func Call[Cl process.Message, R process.Message, S process.Sendable](to S, from process.PID, msg Cl, timeout time.Duration) (resp R, replied bool) {
	if timeout <= 0 {
		timeout = DEFAULT_TIMEOUT
	}
	call := CallMsg[Cl, R](from, msg)
	process.Send(to, call)
	select {
	case resp, ok := <-call.resp:
		return resp, ok
	case <-time.After(timeout):
		return resp, false
	}
}

func Cast[Cl process.Message, S process.Sendable](to S, msg Cl) {
	process.Send(to, CastMsg(msg))
}

// func Stop[S process.Sendable](to S, reason error) {
// 	process.Send(to, StopMsg(reason))
// }
