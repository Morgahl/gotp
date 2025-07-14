package server

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/debug"
	"github.com/Morgahl/gotp/process"
)

// TODO: these shoudl conform once we have a supervisr for them to conform to
// var _ supervisor.Supervisable = &Server[any, any, any, any, any]{}
// var _ supervisor.Supervised = &Server[any, any, any, any, any]{}

type Server[
	I any,
	Cl process.Message,
	R process.Message,
	Cs process.Message,
	Ct process.Message,
] struct {
	server  Serverable[I, Cl, R, Cs, Ct]
	process *process.Process
	initArg I
}

func Start[
	I any,
	Cl process.Message,
	R process.Message,
	Cs process.Message,
	Ct process.Message,
](server Serverable[I, Cl, R, Cs, Ct], initArg I, opts ...process.SpawnOpt) (process.Started, error) {
	return New(server, initArg).Start(opts...)
}

func StartLink[
	I any,
	Cl process.Message,
	R process.Message,
	Cs process.Message,
	Ct process.Message,
](server Serverable[I, Cl, R, Cs, Ct], initArg I, linked *process.Process, opts ...process.SpawnOpt) (Supervised, error) {
	return New(server, initArg).StartLink(linked, opts...)
}

func New[
	I any,
	Cl process.Message,
	R process.Message,
	Cs process.Message,
	Ct process.Message,
](server Serverable[I, Cl, R, Cs, Ct], initArg I) *Server[I, Cl, R, Cs, Ct] {
	return &Server[I, Cl, R, Cs, Ct]{server: server, initArg: initArg}
}

func (s Server[I, Cl, R, Cs, Ct]) String() string {
	return fmt.Sprintf("Server[%T](server: %s, process: %s, initArg: %v)", s.server, s.server, s.process.PID(), s.initArg)
}

func (s *Server[I, Cl, R, Cs, Ct]) ChildSpec() ChildSpec {
	return s.server.ChildSpec()
}

func (s *Server[I, Cl, R, Cs, Ct]) Start(opts ...process.SpawnOpt) (process.Started, error) {
	err := <-s.setupProc(opts...)
	return s, err
}

func (s *Server[I, Cl, R, Cs, Ct]) StartLink(linked *process.Process, opts ...process.SpawnOpt) (Supervised, error) {
	err := <-s.setupLinkedProc(linked, opts...)
	return s, err
}

func (s *Server[I, Cl, R, Cs, Ct]) Call(msg Cl, timeout time.Duration) (resp R, err error) {
	call := CallMsg[Cl, R](s.process.PID(), msg)
	process.Send(s.process, call)

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
	process.Send(s.process, CastMsg(msg))
}

func (s *Server[I, Cl, R, Cs, Ct]) Info(msg process.Message) {
	process.Send(s.process, msg)
}

func (s *Server[I, Cl, R, Cs, Ct]) Stop(reason error) {
	process.Send(s.process, StopMsg(reason))
}

func (s *Server[I, Cl, R, Cs, Ct]) PID() process.PID {
	return s.process.PID()
}

func (s *Server[I, Cl, R, Cs, Ct]) Send(msg process.Message) {
	process.Send(s.process, msg)
}

func (s *Server[I, Cl, R, Cs, Ct]) SendAfter(msg process.Message, after time.Duration) *time.Timer {
	return process.SendAfter(s.process, msg, after)
}

func (s *Server[I, Cl, R, Cs, Ct]) Exit(reason error) {
	s.process.Exit(reason)
}

func (s *Server[I, Cl, R, Cs, Ct]) Process() *process.Process {
	return s.process
}

func (s *Server[I, Cl, R, Cs, Ct]) setupProc(opts ...process.SpawnOpt) <-chan error {
	sig := make(chan error)
	s.process = process.Spawn(s.loop(sig), opts...)
	s.process.Start()
	return sig
}

func (s *Server[I, Cl, R, Cs, Ct]) setupLinkedProc(linked *process.Process, opts ...process.SpawnOpt) <-chan error {
	sig := make(chan error)
	s.process = process.SpawnLink(s.loop(sig), linked, opts...)
	s.process.Start()
	return sig
}

func (s *Server[I, Cl, R, Cs, Ct]) loop(sig chan error) process.RunFn {
	return func(p *process.Process) (reason error) {
		var cont Continue[Ct]
		var resp Response[R]
		defer func() {
			reason = debug.Recover(recover(), "Server.loop", reason)
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
				slog.Debug("Server.loop: exiting due to reason", slog.Any("pid", p.PID()), slog.Any("reason", reason))
				return
			} else if cont.atom == CONTINUE {
				// We have a continuation, we should process it first and then continue the loop.
				cont, reason = s.server.HandleContinue(cont.arg)
				continue
			}

			msg, ok, err := process.ReceiveWithTimeout[process.Message](p, 0)
			if err != nil {
				return err
			} else if !ok {
				debug.Throw("Server.loop: Process message queue closed unexpectedly for PID %s", p.PID())
			}
			slog.Debug("Server.loop: received message", slog.Any("pid", p.PID()), slog.Any("message", msg))
			switch msg := msg.(type) {
			case stop:
				slog.Debug("Server.loop: received stop message", slog.Any("pid", p.PID()), slog.Any("reason", msg.reason))
				// We have a stop message, we should terminate the server.
				return msg.reason

			case process.ExitMsg:
				slog.Debug("Server.loop: received exit message", slog.Any("pid", p.PID()), slog.Any("reason", msg.Reason))
				// We have an Exit message
				cont, reason = s.server.HandleInfo(msg)

			case call[Cl, R]:
				slog.Debug("Server.loop: received call message", slog.Any("pid", p.PID()), slog.Any("message", msg))
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
				slog.Debug("Server.loop: received cast message", slog.Any("pid", p.PID()), slog.Any("message", msg))
				// We have an asynchronous call
				cont, reason = s.server.HandleCast(msg.req)

			default:
				slog.Debug("Server.loop: received unknown message", slog.Any("pid", p.PID()), slog.Any("message", msg))
				// We have an Info or some other message that we don't know how to handle.
				cont, reason = s.server.HandleInfo(msg)
			}
		}
	}
}
