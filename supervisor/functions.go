package supervisor

import (
	"github.com/Morgahl/gotp/dbg"
	"github.com/Morgahl/gotp/gen_server"
	"github.com/Morgahl/gotp/process"
	"github.com/Morgahl/gotp/term"
)

func StartChild[S process.Sendable](supervisor S, child Supervisable) term.Term {
	if resp, ok := gen_server.Call[ChildSpec, term.Term](supervisor, process.PID{}, child.ChildSpec(), 0); ok {
		switch resp := resp.(type) {
		case AlreadyStarted, process.PID, error:
			return resp
		default:
			dbg.Throw("StartChild: unexpected response type %T", resp)
		}
	}
	return nil
}

func StopChild[S process.Sendable](supervisor S, pid process.PID) term.Term {
	if resp, ok := gen_server.Call[process.PID, term.Term](supervisor, process.PID{}, pid, 0); ok {
		switch resp := resp.(type) {
		case bool:
		default:
			dbg.Throw("StopChild: unexpected response type %T", resp)
		}
	}
	return nil
}
