package gen_server

import (
	"time"

	"github.com/Morgahl/gotp"
)

func GenCall[Cl gotp.Msg, R gotp.Msg](to, from gotp.PID, msg Cl, timeout time.Duration) (resp R, replied bool, err error) {
	call := CallMsg[Cl, R](msg, from)
	if err = gotp.Send(to, call, timeout); err != nil {
		return resp, false, err
	}

	if timeout <= 0 {
		timeout = gotp.DEFAULT_TIMEOUT
	}

	select {
	case <-time.After(timeout):
		return resp, false, gotp.NewTimeout(timeout)
	case resp, ok := <-call.resp:
		return resp, ok, nil
	}
}

func GenCast[Cl gotp.Msg](to gotp.PID, msg Cl, timeout time.Duration) error {
	return gotp.Send(to, CastMsg(msg), timeout)
}
