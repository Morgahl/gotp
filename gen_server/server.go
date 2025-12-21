package gen_server

import (
	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/dbg"
	"github.com/Morgahl/gotp/process"
)

type server[
	I any,
	Cl gotp.Term,
	R gotp.Term,
	Cs gotp.Term,
	Ct gotp.Term,
] struct {
	serverable GenServer[I, Cl, R, Cs, Ct]
}

func (s *server[I, Cl, R, Cs, Ct]) setupProc(initArg I, opts ...process.SpawnOpt) (process.Ref, <-chan error) {
	sig := make(chan error)
	ref, err := process.Spawn(s.loop(initArg, sig), opts...)
	if err != nil {
		sig <- err
		close(sig)
		return process.Ref{}, sig
	}
	return ref, sig
}

func (s *server[I, Cl, R, Cs, Ct]) setupLinkedProc(initArg I, linked process.Ref, opts ...process.SpawnOpt) (process.Ref, <-chan error) {
	sig := make(chan error)
	ref, err := process.SpawnLink(s.loop(initArg, sig), linked, opts...)
	if err != nil {
		sig <- err
		close(sig)
		return process.Ref{}, sig
	}
	return ref, sig
}

func (s *server[I, Cl, R, Cs, Ct]) loop(initArg I, sig chan error) process.RunFn {
	return func(pctx process.Context) (reason error) {
		var cont Continue[Ct]
		var resp Response[R]
		defer func() {
			reason = dbg.Recover(recover(), "server.loop", reason)
			reason = s.serverable.Terminate(pctx, reason)
			if sig != nil && len(sig) < cap(sig) {
				sig <- reason
				close(sig)
			}
		}()

		cont, reason = s.serverable.Init(pctx, initArg)
		sig <- reason
		close(sig)
		sig = nil

		for {
			if reason != nil {
				// stop processing messages
				return
			} else if cont.atom == STOP {
				// We have a stop message, we should terminate the server.
				reason = any(cont.arg).(error)
				return
			} else if cont.atom == CONTINUE {
				// We have a continuation, we should process it first and then continue the loop.
				cont, reason = s.serverable.HandleContinue(pctx, cont.arg)
				continue
			}

			msg, ok, err := process.ReceiveWithTimeout[gotp.Term](pctx, 0)
			if err != nil {
				return err
			} else if !ok {
				dbg.Throw("server.loop: Process message queue closed unexpectedly for PID %s", pctx.PID())
			}
			switch msg := msg.(type) {
			case stop:
				// We have a stop message, we should terminate the server.
				return msg.reason

			case process.ExitMsg:
				// We have an Exit message
				cont, reason = s.serverable.HandleInfo(pctx, msg)

			case call[Cl, R]:
				// We have a synchronous call and a chan to close after conditionally sending a
				// response back to the caller.
				resp, cont, reason = s.serverable.HandleCall(pctx, msg.req, msg.from)
				if reason == nil {
					switch resp._type {
					case NO_REPLY:
						// We have been asked to not send a response back to the caller so just close
						// the resp chan.
						close(msg.resp)

					case REPLY:
						// We have been asked to send a response back to the caller.
						msg.resp <- resp.resp
						close(msg.resp)
					}
				}

			case cast[Cs]:
				// We have an asynchronous call
				cont, reason = s.serverable.HandleCast(pctx, msg.req)

			default:
				// We have an Info or some other message that we don't know how to handle.
				cont, reason = s.serverable.HandleInfo(pctx, msg)
			}
		}
	}
}
