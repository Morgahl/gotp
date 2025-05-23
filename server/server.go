package server

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/supervisor"
)

type Serverable[
	Call gotp.Msg,
	Resp gotp.Msg,
	Cast gotp.Msg,
	Info gotp.Msg,
	Cont any,
] interface {
	supervisor.Supervisable
	Init(supervisor.Options) (Continue[Cont], error)
	HandleCall(Call, gotp.PID) (Response[Resp], Continue[Cont], error)
	HandleCast(Cast) (Continue[Cont], error)
	HandleContinue(Cont) (Continue[Cont], error)
	HandleInfo(Info) (Continue[Cont], error)
	HandleAny(gotp.Msg) (Continue[Cont], error)
	Terminate(error) error
}

type Server[
	Call gotp.Msg,
	Resp gotp.Msg,
	Cast gotp.Msg,
	Info gotp.Msg,
	Cont any,
] struct {
	server  Serverable[Call, Resp, Cast, Info, Cont]
	process *gotp.Process
	conf    supervisor.Options
}

func Start[
	Call gotp.Msg,
	Resp gotp.Msg,
	Cast gotp.Msg,
	Info gotp.Msg,
	Cont any,
](server Serverable[Call, Resp, Cast, Info, Cont], timeout time.Duration, opts ...gotp.SpawnOpt) (supervisor.Supervisable, error) {
	slog.Debug("Server.Start", slog.String("opts", fmt.Sprintf("%v", opts)))
	return New(server).Start(timeout, opts...)
}

func StartLink[
	Call gotp.Msg,
	Resp gotp.Msg,
	Cast gotp.Msg,
	Info gotp.Msg,
	Cont any,
](link gotp.PID, server Serverable[Call, Resp, Cast, Info, Cont], timeout time.Duration, opts ...gotp.SpawnOpt) (supervisor.Supervisable, error) {
	slog.Debug("Server.StartLink", slog.String("link", link.String()))
	return New(server).StartLink(link, timeout, opts...)
}

func New[
	Call gotp.Msg,
	Resp gotp.Msg,
	Cast gotp.Msg,
	Info gotp.Msg,
	Cont any,
](server Serverable[Call, Resp, Cast, Info, Cont]) *Server[Call, Resp, Cast, Info, Cont] {
	return &Server[Call, Resp, Cast, Info, Cont]{server: server}
}

func (s *Server[Call, Resp, Cast, Info, Cont]) Start(timeout time.Duration, opts ...gotp.SpawnOpt) (supervisor.Supervisable, error) {
	slog.Debug("Server.Start", slog.String("opts", fmt.Sprintf("%v", opts)))

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

func (s *Server[Call, Resp, Cast, Info, Cont]) StartLink(link gotp.PID, timeout time.Duration, opts ...gotp.SpawnOpt) (supervisor.Supervisable, error) {
	slog.Debug("Server.StartLink", slog.String("link", link.String()))
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

func (s *Server[Call, Resp, Cast, Info, Cont]) ID() gotp.PID {
	return s.process.ID()
}

func (s *Server[Call, Resp, Cast, Info, Cont]) ChildSpec() supervisor.ChildSpec {
	return s.server.ChildSpec()
}

func (s *Server[Call, Resp, Cast, Info, Cont]) Call(msg Call, timeout time.Duration) (resp Resp, err error) {
	slog.Debug("Server.Call", slog.String("msg", fmt.Sprintf("%v", msg)))
	call := newCallMsg[Call, Resp](msg, s.process.ID())
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

func (s *Server[Call, Resp, Cast, Info, Cont]) Cast(msg Cast, timeout time.Duration) error {
	slog.Debug("Server.Cast", slog.String("msg", fmt.Sprintf("%v", msg)))
	return s.process.Send(newCastMsg(msg), timeout)
}

func (s *Server[Call, Resp, Cast, Info, Cont]) Info(msg Info, timeout time.Duration) error {
	slog.Debug("Server.Info", slog.String("msg", fmt.Sprintf("%v", msg)))
	return s.process.Send(newInfoMsg(msg), timeout)
}

func (s *Server[Call, Resp, Cast, Info, Cont]) Exit(reason error, timeout time.Duration) error {
	slog.Error("Server.Exit", "reason", reason)
	return s.process.Send(gotp.NewExit(s.ID(), reason), timeout)
}

func (s *Server[Call, Resp, Cast, Info, Cont]) Exited() bool {
	slog.Debug("Server.Exited")
	return s.process.Exited()
}

func (s *Server[Call, Resp, Cast, Info, Cont]) Send(msg gotp.Msg, timeout time.Duration) error {
	slog.Debug("Server.Send", slog.String("msg", fmt.Sprintf("%v", msg)))
	if s.process == nil {
		return errors.New("Server not started")
	}
	return s.process.Send(msg, timeout)
}

func (s *Server[Call, Resp, Cast, Info, Cont]) setupProc(link gotp.PID, timeout time.Duration, opts ...gotp.SpawnOpt) <-chan struct{} {
	sig := make(chan struct{})
	s.process = gotp.SpawnLink(link, s.loop(sig), timeout, opts...)
	return sig
}

func (s *Server[Call, Resp, Cast, Info, Cont]) loop(sig chan struct{}) gotp.RunFn {
	return func(_ *gotp.Process, in <-chan gotp.Msg) (reason error) {
		var cont Continue[Cont]
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

		slog.Debug("Server.loop: before init")

		cont, reason = s.server.Init(s.conf)
		close(sig)

		slog.Debug("Server.loop: after init")
		for {
			if reason != nil {
				return // stop processing messages
			} else if cont.atom == CONTINUE {
				slog.Debug("Server.loop: continue")
				// We have a continuation, we should process it first and then continue the loop.
				cont, reason = s.server.HandleContinue(cont.arg)
				continue
			}

			for msg := range in {
				slog.Debug("Server.loop: received message", "msg", msg)

				switch msg := msg.(type) {
				case callMsg[Call, Resp]:
					slog.Debug("Server.loop: call")
					// We have a synchronous call and a chan to close after conditionally sending a
					// response back to the caller.
					var resp Response[Resp]
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

				case castMsg[Cast]:
					slog.Debug("Server.loop: cast", "msg", msg)
					// We have an asynchronous call
					cont, reason = s.server.HandleCast(msg.cast)

				case infoMsg[Info]:
					slog.Debug("Server.loop: info", "msg", msg)
					// We have an Info message
					cont, reason = s.server.HandleInfo(msg.info)

				case gotp.Exit:
					// We have an Exit message
					if msg.PID() == s.process.ID() {
						// We have been asked to terminate
						slog.Debug("Server.loop: exit")
						return msg.Unwrap()
					}

				default:
					slog.Debug("Server.loop: default", "msg", msg)
					// We have an Info or some other message that we don't know how to handle.
					cont, reason = s.server.HandleAny(msg)
				}
			}
		}
	}
}
