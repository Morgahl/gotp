package supervisor

import (
	"github.com/Morgahl/gotp/debug"
	"github.com/Morgahl/gotp/gen_server"
	"github.com/Morgahl/gotp/process"
)

func StartChild[S process.Sendable](supervisor S, child Supervisable) process.Message {
	if resp, ok := gen_server.Call[ChildSpec, process.Message](supervisor, process.PID{}, child.ChildSpec(), 0); ok {
		switch resp := resp.(type) {
		case AlreadyStarted, process.PID, error:
			return resp
		default:
			debug.Throw("StartChild: unexpected response type %T", resp)
		}
	}
	return nil
}

func ContextStartChild[S process.Sendable](pctx process.Context, supervisor S, child Supervisable) process.Message {
	if resp, ok := gen_server.ContextCall[ChildSpec, process.Message, S](supervisor, pctx, child.ChildSpec(), 0); ok {
		switch resp := resp.(type) {
		case AlreadyStarted, process.PID, error:
			return resp
		default:
			debug.Throw("ContextStartChild: unexpected response type %T", resp)
		}
	}
	return nil
}

func StopChild[S process.Sendable](supervisor S, pid process.PID) process.Message {
	if resp, ok := gen_server.Call[process.PID, process.Message](supervisor, process.PID{}, pid, 0); ok {
		switch resp := resp.(type) {
		case bool:
		default:
			debug.Throw("StopChild: unexpected response type %T", resp)
		}
	}
	return nil
}
