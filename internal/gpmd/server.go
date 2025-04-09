package gpmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/netip"

	"github.com/Morgahl/gotp/internal/ctx"
)

type Server struct {
	ctx  ctx.Cancellable
	addr netip.AddrPort
	gpmd *GPMD
}

func NewServer(ctx ctx.Cancellable, listen netip.AddrPort, gpmd *GPMD) (*Server, error) {
	server := &Server{
		ctx:  ctx,
		addr: listen,
		gpmd: gpmd,
	}

	if err := server.Listen(); err != nil {
		return nil, err
	}

	return server, nil
}

func (s *Server) Listen() (err error) {
	var listener net.Listener
	listener, err = net.Listen("tcp", s.addr.String())
	if err != nil {
		return err
	}

	go func() {
		defer listener.Close()
		for {
			select {
			case <-s.ctx.Done():
				return
			default:
				conn, err := listener.Accept()
				if err != nil {
					log.Printf("failed to accept connection: %s", err.Error())
					continue
				}

				go s.handleConnection(conn)
			}
		}
	}()

	return nil
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	dec := json.NewDecoder(conn)
	enc := json.NewEncoder(conn)
	for {
		if err := s.checkCtxs(); err != nil {
			return
		}

		var action ActionRequest
		switch err := dec.Decode(&action); err {
		case io.EOF:
			return
		case nil:
		default:
			log.Printf("failed to decode action: %s", err)
			return
		}

		log.Printf("<<< %s", action)

		switch action.ActionType {
		default:
			err := fmt.Errorf("unknown action type: %s", action.ActionType)
			log.Println(err)
			result := ErrorResult[None](action.ActionType, err)
			if err := send(enc, result); err != nil {
				return
			}

		case ActionRegister:
			result := s.handleRegister(action)
			if err := send(enc, result); err != nil {
				return
			}

		case ActionUnregister:
			result := s.handleUnregister(action)
			if err := send(enc, result); err != nil {
				return
			}

		case ActionHeartbeat:
			result := s.handleHeartbeat(action)
			if err := send(enc, result); err != nil {
				return
			}

		case ActionNames:
			result := s.handleNames(action)
			if err := send(enc, result); err != nil {
				return
			}

		case ActionKill:
			result := s.handleKill(action)
			if err := send(enc, result); err != nil {
				return
			}

			if result.Result.Error == nil {
				s.ctx.Cancel(fmt.Errorf("received kill request"))
				return
			}
		}
	}
}

func (s *Server) handleRegister(action ActionRequest) ActionResult[None] {
	if action.Node.Name == "" {
		return ErrorResult[None](action.ActionType, fmt.Errorf("register action missing node"))
	} else if err := s.gpmd.Register(action.Node); err != nil {
		return ErrorResult[None](action.ActionType, err)
	}

	return SuccessResult(action.ActionType, None{})
}

func (s *Server) handleUnregister(action ActionRequest) ActionResult[None] {
	if action.Node.Name == "" {
		return ErrorResult[None](action.ActionType, fmt.Errorf("unregister action missing node"))
	} else if err := s.gpmd.Unregister(action.Node); err != nil {
		return ErrorResult[None](action.ActionType, err)
	}

	return SuccessResult(action.ActionType, None{})
}

func (s *Server) handleHeartbeat(action ActionRequest) ActionResult[None] {
	if action.Node.Name == "" {
		return ErrorResult[None](action.ActionType, fmt.Errorf("heartbeat action missing node"))
	} else if err := s.gpmd.Heartbeat(action.Node); err != nil {
		return ErrorResult[None](action.ActionType, err)
	}

	return SuccessResult(action.ActionType, None{})
}

func (s *Server) handleNames(action ActionRequest) ActionResult[NodeList] {
	nodes, err := s.gpmd.List()
	if err != nil {
		return ErrorResult[NodeList](action.ActionType, err)
	}

	return SuccessResult(action.ActionType, nodes)
}

func (s *Server) handleKill(action ActionRequest) ActionResult[None] {
	nodes, err := s.gpmd.List()
	if err != nil {
		return ErrorResult[None](action.ActionType, err)
	} else if len(nodes) > 0 {
		err = fmt.Errorf("cannot kill as there are still %d nodes registered", len(nodes))
		return ErrorResult[None](action.ActionType, err)
	}

	return SuccessResult(action.ActionType, None{})
}

func (s *Server) checkCtxs() error {
	select {
	case <-s.ctx.Done():
		return context.Cause(s.ctx)
	default:
		return nil
	}
}

func send[R JSONable](enc *json.Encoder, result ActionResult[R]) error {
	if err := enc.Encode(result); err != nil {
		err = fmt.Errorf("failed to encode action result: %w", err)
		log.Println(err)
		return err
	}

	log.Printf(">>> %s", result)
	return nil
}
