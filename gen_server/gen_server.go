package gen_server

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

func (s *Server[I, Cl, R, Cs, Ct]) ChildSpec() gotp.ChildSpec {
	return s.server.ChildSpec()
}

func (s *Server[I, Cl, R, Cs, Ct]) Start(opts ...gotp.SpawnOpt) (gotp.Started, error) {
	<-s.setupProc(opts...)
	return s, nil
}

func (s *Server[I, Cl, R, Cs, Ct]) StartLink(link gotp.PID, opts ...gotp.SpawnOpt) (gotp.Supervised, error) {
	<-s.setupLinkedProc(link, opts...)
	return s, nil
}

func (s *Server[I, Cl, R, Cs, Ct]) Call(msg Cl, timeout time.Duration) (resp R, err error) {
	call := CallMsg[Cl, R](msg, s.process.PID())
	if err = s.process.Send(call, timeout); err != nil {
		return resp, err
	}

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

func (s *Server[I, Cl, R, Cs, Ct]) Cast(msg Cs, timeout time.Duration) error {
	return s.process.Send(CastMsg(msg), timeout)
}

func (s *Server[I, Cl, R, Cs, Ct]) Info(msg gotp.Msg, timeout time.Duration) error {
	return s.process.Send(msg, timeout)
}

func (s *Server[I, Cl, R, Cs, Ct]) PID() gotp.PID {
	return s.process.PID()
}

func (s *Server[I, Cl, R, Cs, Ct]) Send(msg gotp.Msg, timeout time.Duration) error {
	if s.process == nil {
		return gotp.NewNotStarted()
	}
	return s.process.Send(msg, timeout)
}

func (s *Server[I, Cl, R, Cs, Ct]) SendAfter(msg gotp.Msg, delay time.Duration) (*time.Timer, error) {
	if s.process == nil {
		slog.Error("Server.SendAfter: Server not started")
		return nil, gotp.NewNotStarted()
	}
	return s.process.SendAfter(msg, delay)
}

func (s *Server[I, Cl, R, Cs, Ct]) Receive() <-chan gotp.Msg {
	if s.process == nil {
		slog.Error("Server.Receive: Server not started")
		return nil
	}
	return s.process.Receive()
}

func (s *Server[I, Cl, R, Cs, Ct]) setupProc(opts ...gotp.SpawnOpt) <-chan struct{} {
	sig := make(chan struct{})
	s.process = gotp.Spawn(s.loop(sig), opts...)
	return sig
}

func (s *Server[I, Cl, R, Cs, Ct]) setupLinkedProc(link gotp.PID, opts ...gotp.SpawnOpt) <-chan struct{} {
	sig := make(chan struct{})
	s.process = gotp.SpawnLink(s.loop(sig), link, opts...)
	return sig
}

func (s *Server[I, Cl, R, Cs, Ct]) loop(sig chan struct{}) gotp.RunFn {
	return func(p *gotp.Process) (reason error) {
		var cont Continue[Ct]
		var resp Response[R]
		defer func() {
			if r := recover(); r != nil {
				reason = debug.Catch(reason, r)
			}
			if err := s.process.Exit(s.server.Terminate(reason), 0); err != nil {
				slog.Error("Server.loop: failed to exit process", "error", err)
			}
		}()

		cont, reason = s.server.Init(s.initArg)
		close(sig)

		for {
			if reason != nil {
				return // stop processing messages
			} else if cont.atom == CONTINUE {
				// We have a continuation, we should process it first and then continue the loop.
				cont, reason = s.server.HandleContinue(cont.arg)
				continue
			}

			msg, ok := <-s.process.Receive()
			if !ok {
				// The Process mailbox has been closed?!?!
				panic(fmt.Sprintf("Server.loop: Process mailbox closed unexpectedly for PID %s", s.process.PID()))
			}
			switch msg := msg.(type) {
			case gotp.Exit:
				// We have an Exit message
				if msg.PID() == s.process.PID() {
					// We have been asked to terminate
					return msg.Unwrap()
				}

				cont, reason = s.server.HandleInfo(msg)

			case call[Cl, R]:
				// We have a synchronous call and a chan to close after conditionally sending a
				// response back to the caller.
				switch resp, cont, reason = s.server.HandleCall(msg.req, msg.from); resp.atom {
				case NO_REPLY:
					// We have been asked to not send a response back to the caller so just close
					// the resp chan.
					close(msg.resp)

				case REPLY:
					// We have been asked to send a response back to the caller.
					msg.resp <- resp.resp
					close(msg.resp)
				}

			case cast[Cs]:
				// We have an asynchronous call
				cont, reason = s.server.HandleCast(msg.cast)

			default:
				// We have an Info or some other message that we don't know how to handle.
				cont, reason = s.server.HandleInfo(msg)
			}
		}
	}
}
