package gen_server

import (
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/process"
)

const (
	DEFAULT_TIMEOUT = 5 * time.Second
)

func Start[
	I any,
	Cl gotp.Term,
	R gotp.Term,
	Cs gotp.Term,
	Ct gotp.Term,
](serverable GenServer[I, Cl, R, Cs, Ct], initArg I, opts ...process.SpawnOpt) (process.Ref, error) {
	s := server[I, Cl, R, Cs, Ct]{serverable: serverable}
	pid, errCh := s.setupProc(initArg, opts...)
	return pid, <-errCh
}

func StartLink[
	I any,
	Cl gotp.Term,
	R gotp.Term,
	Cs gotp.Term,
	Ct gotp.Term,
](serverable GenServer[I, Cl, R, Cs, Ct], initArg I, linked process.Ref, opts ...process.SpawnOpt) (process.Ref, error) {
	s := server[I, Cl, R, Cs, Ct]{serverable: serverable}
	pid, errCh := s.setupLinkedProc(initArg, linked, opts...)
	return pid, <-errCh
}

func Call[Cl gotp.Term, R gotp.Term, S process.Sendable](to S, from process.PID, msg Cl, timeout time.Duration) (resp R, replied bool) {
	if timeout < 0 {
		timeout = DEFAULT_TIMEOUT
	}
	var after <-chan time.Time
	if timeout > 0 {
		after = time.After(timeout)
	}
	call := CallMsg[Cl, R](from, msg)
	process.Send(to, call)
	select {
	case resp, ok := <-call.resp:
		return resp, ok
	case <-after:
		return resp, false
	}
}

func Cast[Cl gotp.Term, S process.Sendable](to S, msg Cl) {
	process.Send(to, CastMsg(msg))
}
