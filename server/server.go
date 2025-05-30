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
	I gotp.Msg,
	Ct any,
] interface {
	ChildSpec() gotp.ChildSpec
	Init(gotp.Options) (Continue[Ct], error)
	HandleCall(Cl, gotp.PID) (Response[R], Continue[Ct], error)
	HandleCast(Cs) (Continue[Ct], error)
	HandleContinue(Ct) (Continue[Ct], error)
	HandleInfo(I) (Continue[Ct], error)
	HandleAny(gotp.Msg) (Continue[Ct], error)
	Terminate(error) error
}

var _ gotp.Supervisable = &Server[any, any, any, any, any]{}
var _ gotp.Supervised = &Server[any, any, any, any, any]{}

type Server[
	Cl gotp.Msg,
	R gotp.Msg,
	Cs gotp.Msg,
	I gotp.Msg,
	Ct any,
] struct {
	server  Serverable[Cl, R, Cs, I, Ct]
	process *gotp.Process
	conf    gotp.Options
}

func Start[
	Cl gotp.Msg,
	R gotp.Msg,
	Cs gotp.Msg,
	I gotp.Msg,
	Ct any,
](server Serverable[Cl, R, Cs, I, Ct], timeout time.Duration, opts ...gotp.SpawnOpt) (gotp.Started, error) {
	return New(server).Start(timeout, opts...)
}

func StartLink[
	Cl gotp.Msg,
	R gotp.Msg,
	Cs gotp.Msg,
	I gotp.Msg,
	Ct any,
](server Serverable[Cl, R, Cs, I, Ct], link gotp.PID, timeout time.Duration, opts ...gotp.SpawnOpt) (gotp.Supervised, error) {
	return New(server).StartLink(link, timeout, opts...)
}

func New[
	Cl gotp.Msg,
	R gotp.Msg,
	Cs gotp.Msg,
	I gotp.Msg,
	Ct any,
](server Serverable[Cl, R, Cs, I, Ct]) *Server[Cl, R, Cs, I, Ct] {
	return &Server[Cl, R, Cs, I, Ct]{server: server}
}

func (s *Server[Cl, R, Cs, I, Ct]) ChildSpec() gotp.ChildSpec {
	return s.server.ChildSpec()
}

func (s *Server[Cl, R, Cs, I, Ct]) Start(timeout time.Duration, opts ...gotp.SpawnOpt) (gotp.Started, error) {
	if timeout <= 0 {
		timeout = gotp.DEFAULT_TIMEOUT
	}

	select {
	case <-time.After(timeout):
		return nil, gotp.NewTimeout(timeout)
	case <-s.setupProc(gotp.PIDZero(), timeout, opts...):
		return s, nil
	}
}

func (s *Server[Cl, R, Cs, I, Ct]) StartLink(link gotp.PID, timeout time.Duration, opts ...gotp.SpawnOpt) (gotp.Supervised, error) {
	if timeout <= 0 {
		timeout = gotp.DEFAULT_TIMEOUT
	}

	select {
	case <-time.After(timeout):
		return nil, gotp.NewTimeout(timeout)
	case <-s.setupProc(link, timeout, opts...):
		return s, nil
	}
}

func (s *Server[Cl, R, Cs, I, Ct]) PID() gotp.PID {
	return s.process.PID()
}

func (s *Server[Cl, R, Cs, I, Ct]) Call(msg Cl, timeout time.Duration) (resp R, err error) {
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

func (s *Server[Cl, R, Cs, I, Ct]) Cast(msg Cs, timeout time.Duration) error {
	return s.process.Send(Cast(msg), timeout)
}

func (s *Server[Cl, R, Cs, I, Ct]) Info(msg I, timeout time.Duration) error {
	return s.process.Send(newInfoMsg(msg), timeout)
}

func (s *Server[Cl, R, Cs, I, Ct]) Exit(reason error, timeout time.Duration) error {
	return s.process.Send(gotp.NewExit(s.PID(), reason), timeout)
}

func (s *Server[Cl, R, Cs, I, Ct]) Exited() bool {
	return s.process.Exited()
}

func (s *Server[Cl, R, Cs, I, Ct]) Send(msg gotp.Msg, timeout time.Duration) error {
	if s.process == nil {
		return gotp.NewNotStarted()
	}
	return s.process.Send(msg, timeout)
}

func (s *Server[Cl, R, Cs, I, Ct]) SendAfter(msg gotp.Msg, delay time.Duration) (*time.Timer, error) {
	if s.process == nil {
		slog.Error("Server.SendAfter: Server not started")
		return nil, gotp.NewNotStarted()
	}
	return s.process.SendAfter(msg, delay)
}

func (s *Server[Cl, R, Cs, I, Ct]) Receive() <-chan gotp.Msg {
	if s.process == nil {
		slog.Error("Server.Receive: Server not started")
		return nil
	}
	return s.process.Receive()
}

func (s *Server[Cl, R, Cs, I, Ct]) setupProc(link gotp.PID, timeout time.Duration, opts ...gotp.SpawnOpt) <-chan struct{} {
	sig := make(chan struct{})
	s.process = gotp.SpawnLink(link, s.loop(sig), timeout, opts...)
	return sig
}

func (s *Server[Cl, R, Cs, I, Ct]) loop(sig chan struct{}) gotp.RunFn {
	return func(_ *gotp.Process, in <-chan gotp.Msg) (reason error) {
		var cont Continue[Ct]
		defer func() {
			if r := recover(); r != nil {
				if reason == nil {
					reason = fmt.Errorf("panic: %v", r)
				} else {
					reason = fmt.Errorf("reason: %w, panic: %v", reason, r)
				}
			}
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

			for msg := range in {
				switch msg := msg.(type) {
				case call[Cl, R]:
					// We have a synchronous call and a chan to close after conditionally sending a
					// response back to the caller.
					var resp Response[R]
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

				case infoMsg[I]:
					// We have an Info message
					cont, reason = s.server.HandleInfo(msg.info)

				case gotp.Exit:
					// We have an Exit message
					if msg.PID() == s.process.PID() {
						// We have been asked to terminate
						return msg.Unwrap()
					}

				default:
					// We have an Info or some other message that we don't know how to handle.
					cont, reason = s.server.HandleAny(msg)
				}
			}
		}
	}
}
