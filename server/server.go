package server

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/Morgahl/gotp"
)

type Serverable[
	Cl gotp.Msg,
	R gotp.Msg,
	Cs gotp.Msg,
	Ct gotp.Msg,
] interface {
	ChildSpec() gotp.ChildSpec
	Init(gotp.Options) (Continue[Ct], error)
	HandleCall(Cl, gotp.PID) (Response[R], Continue[Ct], error)
	HandleCast(Cs) (Continue[Ct], error)
	HandleContinue(Ct) (Continue[Ct], error)
	HandleInfo(gotp.Msg) (Continue[Ct], error)
	Terminate(error) error
}

var _ gotp.Supervisable = &Server[any, any, any, any]{}
var _ gotp.Supervised = &Server[any, any, any, any]{}

type Server[
	Cl gotp.Msg,
	R gotp.Msg,
	Cs gotp.Msg,
	Ct gotp.Msg,
] struct {
	server  Serverable[Cl, R, Cs, Ct]
	process *gotp.Process
	conf    gotp.Options
}

func Start[
	Cl gotp.Msg,
	R gotp.Msg,
	Cs gotp.Msg,
	Ct gotp.Msg,
](server Serverable[Cl, R, Cs, Ct], timeout time.Duration, opts ...gotp.SpawnOpt) (gotp.Started, error) {
	return New(server).Start(timeout, opts...)
}

func StartLink[
	Cl gotp.Msg,
	R gotp.Msg,
	Cs gotp.Msg,
	Ct gotp.Msg,
](server Serverable[Cl, R, Cs, Ct], link gotp.PID, timeout time.Duration, opts ...gotp.SpawnOpt) (gotp.Supervised, error) {
	return New(server).StartLink(link, timeout, opts...)
}

func New[
	Cl gotp.Msg,
	R gotp.Msg,
	Cs gotp.Msg,
	Ct gotp.Msg,
](server Serverable[Cl, R, Cs, Ct]) *Server[Cl, R, Cs, Ct] {
	return &Server[Cl, R, Cs, Ct]{server: server}
}

func (s *Server[Cl, R, Cs, Ct]) ChildSpec() gotp.ChildSpec {
	return s.server.ChildSpec()
}

func (s *Server[Cl, R, Cs, Ct]) Start(timeout time.Duration, opts ...gotp.SpawnOpt) (gotp.Started, error) {
	return s.StartLink(gotp.PIDZero(), timeout, opts...)
}

func (s *Server[Cl, R, Cs, Ct]) StartLink(link gotp.PID, timeout time.Duration, opts ...gotp.SpawnOpt) (gotp.Supervised, error) {
	if timeout <= 0 {
		timeout = gotp.DEFAULT_TIMEOUT
	}

	select {
	case <-time.After(timeout):
		return nil, gotp.NewTimeout(timeout)
	case <-s.setupProc(link, opts...):
		return s, nil
	}
}

func (s *Server[Cl, R, Cs, Ct]) Call(msg Cl, timeout time.Duration) (resp R, err error) {
	call := Call[Cl, R](msg, s.process.PID())
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

func (s *Server[Cl, R, Cs, Ct]) Cast(msg Cs, timeout time.Duration) error {
	return s.process.Send(Cast(msg), timeout)
}

func (s *Server[Cl, R, Cs, Ct]) Info(msg gotp.Msg, timeout time.Duration) error {
	return s.process.Send(msg, timeout)
}

func (s *Server[Cl, R, Cs, Ct]) PID() gotp.PID {
	return s.process.PID()
}

func (s *Server[Cl, R, Cs, Ct]) Send(msg gotp.Msg, timeout time.Duration) error {
	if s.process == nil {
		return gotp.NewNotStarted()
	}
	return s.process.Send(msg, timeout)
}

func (s *Server[Cl, R, Cs, Ct]) SendAfter(msg gotp.Msg, delay time.Duration) (*time.Timer, error) {
	if s.process == nil {
		slog.Error("Server.SendAfter: Server not started")
		return nil, gotp.NewNotStarted()
	}
	return s.process.SendAfter(msg, delay)
}

func (s *Server[Cl, R, Cs, Ct]) Receive() <-chan gotp.Msg {
	if s.process == nil {
		slog.Error("Server.Receive: Server not started")
		return nil
	}
	return s.process.Receive()
}

func (s *Server[Cl, R, Cs, Ct]) setupProc(link gotp.PID, opts ...gotp.SpawnOpt) <-chan struct{} {
	sig := make(chan struct{})
	s.process = gotp.SpawnLink(s.loop(sig), link, opts...)
	return sig
}

func (s *Server[Cl, R, Cs, Ct]) loop(sig chan struct{}) gotp.RunFn {
	return func(p *gotp.Process) (reason error) {
		var cont Continue[Ct]
		var resp Response[R]
		defer func() {
			gotp.Recover(&reason)
			if err := s.process.Exit(s.server.Terminate(reason), 0); err != nil {
				slog.Error("Server.loop: failed to exit process", "error", err)
			}
		}()

		cont, reason = s.server.Init(s.conf)
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
