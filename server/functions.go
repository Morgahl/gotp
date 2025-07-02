package server

import (
	"github.com/Morgahl/gotp/process"
)

func Call[Cl process.Message, R process.Message, S process.Sendable](to S, from process.PID, msg Cl) (resp R, replied bool) {
	call := CallMsg[Cl, R](from, msg)
	process.Send(to, call)
	resp, ok := <-call.resp
	return resp, ok
}

func Cast[Cl process.Message, S process.Sendable](to S, msg Cl) {
	process.Send(to, CastMsg(msg))
}

func Stop[S process.Sendable](to S, reason error) {
	process.Send(to, StopMsg(reason))
}
