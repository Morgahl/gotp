package server

import (
	"context"
	"fmt"
	"log"

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
](ctx context.Context, server Serverable[Call, Resp, Cast, Info, Cont], opts ...gotp.SpawnOpt) *Server[Call, Resp, Cast, Info, Cont] {
	log.Printf("Start: %v", opts)
	return New(server).Start(ctx, opts...)
}

func StartLink[
	Call gotp.Msg,
	Resp gotp.Msg,
	Cast gotp.Msg,
	Info gotp.Msg,
	Cont any,
](ctx context.Context, link gotp.PID, server Serverable[Call, Resp, Cast, Info, Cont], opts ...gotp.SpawnOpt) (supervisor.Supervisable, error) {
	log.Printf("StartLink: %v", link)
	return New(server).StartLink(ctx, link, opts...)
}

func New[
	Call gotp.Msg,
	Resp gotp.Msg,
	Cast gotp.Msg,
	Info gotp.Msg,
	Cont any,
](server Serverable[Call, Resp, Cast, Info, Cont]) *Server[Call, Resp, Cast, Info, Cont] {
	log.Printf("New: %v", server)
	return &Server[Call, Resp, Cast, Info, Cont]{server: server}
}

func (s *Server[Call, Resp, Cast, Info, Cont]) Start(ctx context.Context, opts ...gotp.SpawnOpt) *Server[Call, Resp, Cast, Info, Cont] {
	log.Printf("Server.Start: %v", opts)
	s.setupProc(ctx, gotp.PIDZero(), opts...)
	return s
}

func (s *Server[Call, Resp, Cast, Info, Cont]) StartLink(ctx context.Context, link gotp.PID, opts ...gotp.SpawnOpt) (supervisor.Supervisable, error) {
	log.Printf("Server.StartLink: %v", link)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-s.setupProc(ctx, link, opts...):
		return s, nil
	}
}

func (s *Server[Call, Resp, Cast, Info, Cont]) Stop(ctx context.Context, reason error) error {
	return s.process.Send(ctx, gotp.NewExit(s.ID(), reason))
}

func (s *Server[Call, Resp, Cast, Info, Cont]) ID() gotp.PID {
	return s.process.ID()
}

func (s *Server[Call, Resp, Cast, Info, Cont]) ChildSpec() supervisor.ChildSpec {
	return s.server.ChildSpec()
}

func (s *Server[Call, Resp, Cast, Info, Cont]) Call(ctx context.Context, msg Call) (resp Resp, err error) {
	log.Printf("Server.Call: %v", msg)
	call := newCallMsg[Call, Resp](msg, s.process.ID())
	if err = s.process.Send(ctx, call); err != nil {
		return resp, err
	}

	select {
	case <-ctx.Done():
		return resp, gotp.NewTimeout(ctx.Err())

	case resp := <-call.resp:
		return resp, nil
	}
}

func (s *Server[Call, Resp, Cast, Info, Cont]) Cast(ctx context.Context, msg Cast) error {
	log.Printf("Server.Cast: %v", msg)
	return s.process.Send(ctx, newCastMsg(msg))
}

func (s *Server[Call, Resp, Cast, Info, Cont]) Info(ctx context.Context, msg Info) error {
	log.Printf("Server.Info: %v", msg)
	return s.process.Send(ctx, newInfoMsg(msg))
}

func (s *Server[Call, Resp, Cast, Info, Cont]) setupProc(ctx context.Context, link gotp.PID, opts ...gotp.SpawnOpt) <-chan struct{} {
	sig := make(chan struct{})
	s.process = gotp.SpawnLink(ctx, link, s.loop(sig), opts...)
	return sig
}

// We reimplement the loop function using Mailbox.Receive(match, ...) Mailbox.Chan no longer exists and should not be used.
func (s *Server[Call, Resp, Cast, Info, Cont]) loop(sig chan struct{}) gotp.RunFn {
	return func(ctx context.Context, p *gotp.Process) (reason error) {
		var cont Continue[Cont]
		defer func() {
			if r := recover(); r != nil {
				if reason == nil {
					reason = fmt.Errorf("panic: %v", r)
				} else {
					reason = fmt.Errorf("reason: %w, panic: %v", reason, r)
				}
			}
			reason = s.server.Terminate(reason)
			s.process.Exit(ctx, reason)
		}()

		log.Printf("Server.loop: before init")

		cont, reason = s.server.Init(s.conf)
		close(sig)

		log.Printf("Server.loop: after init")
		for {
			if reason != nil {
				return // stop processing messages
			}

			select {
			case <-ctx.Done():
				log.Printf("Server.loop: context done")
				// The context has been cancelled, we should stop processing messages and exit normally.
				return

			default:
				if cont.atom == CONTINUE {
					log.Printf("Server.loop: continue")
					// We have a continuation, we should process it first and then continue the loop.
					cont, reason = s.server.HandleContinue(cont.arg)
					continue
				}

				if msg, ok := p.Receive(matchCall[Call, Resp]); ok {
					// We have a synchronous call and a chan to close after conditionally sending a
					// response back to the caller.
					log.Printf("Server.loop: call")
					msg := msg.(callMsg[Call, Resp])
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
				} else if msg, ok := p.Receive(matchCast[Cast]); ok {
					// We have an asynchronous call
					log.Printf("Server.loop: cast")
					msg := msg.(castMsg[Cast])
					cont, reason = s.server.HandleCast(msg.cast)
				} else if msg, ok := p.Receive(matchInfo[Info]); ok {
					// We have an Info message
					log.Printf("Server.loop: info")
					msg := msg.(infoMsg[Info])
					cont, reason = s.server.HandleInfo(msg.info)
				} else if msg, ok := p.Receive(matchExit); ok {
					// We have an Exit message
					log.Printf("Server.loop: exit")
					msg := msg.(gotp.Exit)
					if msg.PID() == s.process.ID() {
						// We have been asked to terminate
						return msg.Unwrap()
					}
				} else if msg, ok := p.Receive(matchAny); ok {
					// We have an Info or some other message that we don't know how to handle.
					log.Printf("Server.loop: default")
					cont, reason = s.server.HandleAny(msg)
				} else if ctx.Err() != nil {
					log.Printf("Server.loop: context done")
					// The context has been cancelled, we should stop processing messages and exit normally.
					return ctx.Err()
				} else {
					log.Printf("Server.loop: no message")
					// No message received, continue the loop.
					continue
				}
			}
		}
	}
}

func matchCall[Call gotp.Msg, Resp gotp.Msg](msg gotp.Msg) bool {
	if _, ok := msg.(callMsg[Call, Resp]); ok {
		return true
	}
	return false
}

func matchCast[Cast gotp.Msg](msg gotp.Msg) bool {
	if _, ok := msg.(castMsg[Cast]); ok {
		return true
	}
	return false
}

func matchInfo[Info gotp.Msg](msg gotp.Msg) bool {
	if _, ok := msg.(infoMsg[Info]); ok {
		return true
	}
	return false
}

func matchExit(msg gotp.Msg) bool {
	if _, ok := msg.(gotp.Exit); ok {
		return true
	}
	return false
}

func matchAny(msg gotp.Msg) bool {
	return true
}
