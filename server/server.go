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
	s.setupProc(ctx, gotp.UNLINKED, opts...)
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

func (s *Server[Call, Resp, Cast, Info, Cont]) loop(sig chan struct{}) func(context.Context, *gotp.Process, <-chan gotp.Msg) error {
	return func(ctx context.Context, _ *gotp.Process, in <-chan gotp.Msg) (reason error) {
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
			}

			select {
			case <-ctx.Done():
				log.Printf("Server.loop: context done")
				// The context has been cancelled, we should stop processing messages and exit shutdown.
				return

			case msg, ok := <-in:
				log.Printf("Server.loop: received message: %T(%v)", msg, msg)
				if !ok {
					log.Printf("Server.loop: channel closed")
					// The channel has been closed, we should stop processing messages and exit normal
					return
				}

				switch msg := msg.(type) {
				case callMsg[Call, Resp]:
					log.Printf("Server.loop: call")
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
					log.Printf("Server.loop: cast")
					// We have an asynchronous call
					cont, reason = s.server.HandleCast(msg.cast)

				case infoMsg[Info]:
					log.Printf("Server.loop: info")
					// We have an Info message
					cont, reason = s.server.HandleInfo(msg.info)

				case gotp.Exit:
					log.Printf("Server.loop: exit")
					// We have an Exit message
					if msg.PID() == s.process.ID() {
						// We have been asked to terminate
						return msg.Unwrap()
					}

				default:
					log.Printf("Server.loop: default")
					// We have an Info or some other message that we don't know how to handle.
					cont, reason = s.server.HandleAny(msg)
				}
			}
		}
	}
}
