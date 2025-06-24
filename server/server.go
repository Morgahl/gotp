package server

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/debug"
)

var _ gotp.Supervisable = &Server[any, any, any, any, any]{}
var _ gotp.Supervised = &Server[any, any, any, any, any]{}

type Server[
	I any,
	Cl gotp.Msg,
	R gotp.Msg,
	Cs gotp.Msg,
	Ct gotp.Msg,
] struct {
	server  Serverable[I, Cl, R, Cs, Ct]
	process *gotp.Process
	initArg I
}

func Start[
	I any,
	Cl gotp.Msg,
	R gotp.Msg,
	Cs gotp.Msg,
	Ct gotp.Msg,
](server Serverable[I, Cl, R, Cs, Ct], initArg I, opts ...gotp.SpawnOpt) (gotp.Started, error) {
	return New(server, initArg).Start(opts...)
}

func StartLink[
	I any,
	Cl gotp.Msg,
	R gotp.Msg,
	Cs gotp.Msg,
	Ct gotp.Msg,
](server Serverable[I, Cl, R, Cs, Ct], initArg I, link gotp.PID, opts ...gotp.SpawnOpt) (gotp.Supervised, error) {
	return New(server, initArg).StartLink(link, opts...)
}

func New[
	I any,
	Cl gotp.Msg,
	R gotp.Msg,
	Cs gotp.Msg,
	Ct gotp.Msg,
](server Serverable[I, Cl, R, Cs, Ct], initArg I) *Server[I, Cl, R, Cs, Ct] {
	return &Server[I, Cl, R, Cs, Ct]{server: server, initArg: initArg}
}

func (s Server[I, Cl, R, Cs, Ct]) String() string {
	return fmt.Sprintf("Server[%T](server: %s, process: %s, initArg: %v)", s.server, s.server, s.process, s.initArg)
}

func (s *Server[I, Cl, R, Cs, Ct]) ChildSpec() gotp.ChildSpec {
	return s.server.ChildSpec()
}

func (s *Server[I, Cl, R, Cs, Ct]) Start(opts ...gotp.SpawnOpt) (gotp.Started, error) {
	err := <-s.setupProc(opts...)
	return s, err
}

func (s *Server[I, Cl, R, Cs, Ct]) StartLink(link gotp.PID, opts ...gotp.SpawnOpt) (gotp.Supervised, error) {
	err := <-s.setupLinkedProc(link, opts...)
	return s, err
}

func (s *Server[I, Cl, R, Cs, Ct]) Call(msg Cl, timeout time.Duration) (resp R, err error) {
	call := CallMsg[Cl, R](s.process.ID(), msg)
	s.process.Send(call)

	if timeout <= 0 {
		timeout = gotp.DEFAULT_TIMEOUT
	}

	select {
	case <-time.After(timeout):
		return resp, gotp.NewTimeout(timeout)

	case resp := <-call.resp:
		return resp, nil
	}
}

func (s *Server[I, Cl, R, Cs, Ct]) Cast(msg Cs) {
	s.process.Send(CastMsg(msg))
}

func (s *Server[I, Cl, R, Cs, Ct]) Info(msg gotp.Msg) {
	s.process.Send(msg)
}

func (s *Server[I, Cl, R, Cs, Ct]) Stop(reason error) {
	s.process.Send(StopMsg(reason))
}

func (s *Server[I, Cl, R, Cs, Ct]) ID() gotp.PID {
	return s.process.ID()
}

func (s *Server[I, Cl, R, Cs, Ct]) Send(msg gotp.Msg) {
	s.process.Send(msg)
}

func (s *Server[I, Cl, R, Cs, Ct]) SendAfter(msg gotp.Msg, after time.Duration) *time.Timer {
	return s.process.SendAfter(msg, after)
}

func (s *Server[I, Cl, R, Cs, Ct]) Receive() <-chan gotp.Msg {
	if s.process == nil {
		slog.Error("Server.Receive: Server not started")
		return nil
	}
	return s.process.Receive()
}

func (s *Server[I, Cl, R, Cs, Ct]) setupProc(opts ...gotp.SpawnOpt) <-chan error {
	sig := make(chan error)
	s.process = gotp.Spawn(s.loop(sig), opts...)
	s.process.Start()
	return sig
}

func (s *Server[I, Cl, R, Cs, Ct]) setupLinkedProc(link gotp.PID, opts ...gotp.SpawnOpt) <-chan error {
	sig := make(chan error)
	s.process = gotp.Spawn(s.loop(sig), append(opts, gotp.Link(link))...)
	s.process.Start()
	return sig
}

func (s *Server[I, Cl, R, Cs, Ct]) loop(sig chan error) gotp.RunFn {
	return func(p *gotp.Process) (reason error) {
		var cont Continue[Ct]
		var resp Response[R]
		defer func() {
			if r := recover(); r != nil {
				reason = debug.Catch(reason, r)
			}
			s.process.Exit(s.server.Terminate(reason))
			if sig != nil && len(sig) < cap(sig) {
				sig <- reason
				close(sig)
			}
		}()

		cont, reason = s.server.Init(s.initArg)
		sig <- reason
		close(sig)
		sig = nil

		for {
			if reason != nil {
				// stop processing messages
				return
			} else if cont.atom == CONTINUE {
				// We have a continuation, we should process it first and then continue the loop.
				cont, reason = s.server.HandleContinue(cont.arg)
				continue
			}

			msg, ok := <-p.Receive()
			if !ok {
				// The Process mailbox has been closed?!?!
				panic(fmt.Sprintf("Server.loop: Process mailbox closed unexpectedly for PID %s", p.ID()))
			}
			switch msg := msg.(type) {
			case stop:
				// We have a stop message, we should terminate the server.
				return msg.reason

			case gotp.Exit:
				// We have an Exit message
				if msg.ID() == p.ID() {
					// We have been asked to terminate
					return msg.Unwrap()
				}

				cont, reason = s.server.HandleInfo(msg)

			case call[Cl, R]:
				// We have a synchronous call and a chan to close after conditionally sending a
				// response back to the caller.
				resp, cont, reason = s.server.HandleCall(msg.req, msg.from)
				if reason == nil {
					switch resp.atom {
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
				cont, reason = s.server.HandleCast(msg.req)

			default:
				// We have an Info or some other message that we don't know how to handle.
				cont, reason = s.server.HandleInfo(msg)
			}
		}
	}
}
