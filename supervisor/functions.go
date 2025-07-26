package supervisor

import (
	"github.com/Morgahl/gotp/dbg"
	"github.com/Morgahl/gotp/gen_server"
	"github.com/Morgahl/gotp/process"
)

func StartChild[S process.Sendable](supervisor S, child Supervisable) process.Message {
	if resp, ok := gen_server.Call[ChildSpec, process.Message](supervisor, process.PID{}, child.ChildSpec(), 0); ok {
		switch resp := resp.(type) {
		case AlreadyStarted, process.PID, error:
			return resp
		default:
			dbg.Throw("StartChild: unexpected response type %T", resp)
		}
	}
	return nil
}

func ContextStartChild[S process.Sendable](pctx *process.Context, supervisor S, child Supervisable) process.Message {
	if resp, ok := gen_server.ContextCall[ChildSpec, process.Message](supervisor, pctx, child.ChildSpec(), 0); ok {
		switch resp := resp.(type) {
		case AlreadyStarted, process.PID, error:
			return resp
		default:
			dbg.Throw("ContextStartChild: unexpected response type %T", resp)
		}
	}
	return nil
}

func StopChild[S process.Sendable](supervisor S, pid process.PID) process.Message {
	if resp, ok := gen_server.Call[process.PID, process.Message](supervisor, process.PID{}, pid, 0); ok {
		switch resp := resp.(type) {
		case bool:
		default:
			dbg.Throw("StopChild: unexpected response type %T", resp)
		}
	}
	return nil
}

func ContextStopChild[S process.Sendable](pctx *process.Context, supervisor S, pid process.PID) process.Message {
	if resp, ok := gen_server.ContextCall[process.PID, process.Message](supervisor, pctx, pid, 0); ok {
		switch resp := resp.(type) {
		case bool:
		default:
			dbg.Throw("ContextStopChild: unexpected response type %T", resp)
		}
	}
	return nil
}
