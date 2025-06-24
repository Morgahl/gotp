package server

import (
	"github.com/Morgahl/gotp"
)

func Call[Cl gotp.Msg, R gotp.Msg](to, from gotp.PID, msg Cl) (resp R, replied bool) {
	call := CallMsg[Cl, R](from, msg)
	gotp.Send(to, call)
	resp, ok := <-call.resp
	return resp, ok
}

func Cast[Cl gotp.Msg](to gotp.PID, msg Cl) {
	gotp.Send(to, CastMsg(msg))
}
