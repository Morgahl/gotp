package supervisor

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/server"
)

type StaticSupervisor struct {
	flags      Flags
	specs      []gotp.Supervisable
	childNames map[string]gotp.PID
	children   map[gotp.PID]child
	server     *server.Server[gotp.Msg, gotp.Msg, gotp.Msg, gotp.Msg, any]
}

func Static(flags Flags, children ...gotp.Supervisable) *StaticSupervisor {
	sup := &StaticSupervisor{
		flags:      flags.ApplyDefaults(),
		specs:      children,
		childNames: make(map[string]gotp.PID, len(children)),
		children:   make(map[gotp.PID]child, len(children)),
	}
	sup.server = server.New(sup)
	return sup
}

func (s *StaticSupervisor) ChildSpec() gotp.ChildSpec {
	return gotp.ChildSpec{
		Restart:     gotp.PERMANENT,
		Shutdown:    30 * time.Second,
		Type:        gotp.SUPERVISOR,
		Significant: true,
	}
}

func (s *StaticSupervisor) StartLink(link gotp.PID, timeout time.Duration, opts ...gotp.SpawnOpt) (gotp.Supervisable, error) {
	slog.Debug("StaticSupervisor.StartLink: starting server", "link", link, "timeout", timeout, "opts", opts)
	return s.server.StartLink(link, timeout, opts...)
}

func (s *StaticSupervisor) PID() gotp.PID {
	return s.server.PID()
}

func (s *StaticSupervisor) Send(msg gotp.Msg, timeout time.Duration) error {
	return s.server.Send(msg, timeout)
}

func (s *StaticSupervisor) SendAfter(msg gotp.Msg, delay time.Duration) (*time.Timer, error) {
	return s.server.SendAfter(msg, delay)
}

func (s *StaticSupervisor) Receive() <-chan gotp.Msg {
	return s.server.Receive()
}

func (s *StaticSupervisor) Exit(reason error, timeout time.Duration) error {
	return s.server.Exit(reason, timeout)
}

func (s *StaticSupervisor) Exited() bool {
	return s.server.Exited()
}

func (s *StaticSupervisor) StartChild(child gotp.Supervisable, timeout time.Duration) error {
	if child == nil {
		return NewInvalidChild(child)
	}
	return s.server.Send(server.Call[any, any](startChild{child}, s.PID()), timeout)
}

func (s *StaticSupervisor) StopChild(pid gotp.PID, timeout time.Duration) error {
	if pid == gotp.PIDZero() {
		return nil
	}
	return s.server.Send(server.Call[any, any](stopChild{pid}, s.PID()), timeout)
}

func (s *StaticSupervisor) Init(gotp.Options) (cont server.Continue[any], err error) {
	slog.Debug("StaticSupervisor.Init: initializing supervisor", "flags", s.flags, "children", len(s.specs))
	for _, child := range s.specs {
		if child == nil {
			slog.Error("StaticSupervisor.Init: child is nil, skipping")
			continue
		} else if running, err := s.startChild(child, 0); err != nil {
			slog.Error("StaticSupervisor.Init: failed to start child", "error", err)
			return server.NoCont[any](), err
		} else {
			slog.Debug("StaticSupervisor.Init: child started", "pid", running.PID())
			s.registerChild(child, running)
		}
	}
	return server.NoCont[any](), nil
}

func (s *StaticSupervisor) HandleCall(msg gotp.Msg, _ gotp.PID) (resp server.Response[gotp.Msg], cont server.Continue[any], err error) {
	switch m := msg.(type) {
	case startChild:
		if pid, ok := s.findChild(m.child); ok {
			slog.Debug("StaticSupervisor.HandleCall: child already started", "pid", pid)
			return server.Reply[gotp.Msg](gotp.NewAlreadyStarted(pid)), server.NoCont[any](), nil
		} else if running, err := s.startChild(m.child, 0); err != nil {
			slog.Error("StaticSupervisor.HandleCall: failed to start child", "error", err)
			return server.Reply[gotp.Msg](err), server.NoCont[any](), nil
		} else {
			slog.Debug("StaticSupervisor.HandleCall: child started", "pid", running.PID())
			s.registerChild(m.child, running)
			return server.Reply[gotp.Msg](running.PID()), server.NoCont[any](), nil
		}

	case stopChild:
		if child, ok := s.findChildByPID(m.pid); !ok {
			slog.Debug("StaticSupervisor.HandleCall: child not found", "pid", m.pid)
			return server.Reply[gotp.Msg](false), server.NoCont[any](), nil
		} else if err := child.running.Exit(gotp.Kill{}, 0); err != nil {
			slog.Error("StaticSupervisor.HandleCall: failed to stop child", "pid", m.pid, "error", err)
			return server.Reply[gotp.Msg](err), server.NoCont[any](), nil
		} else {
			slog.Debug("StaticSupervisor.HandleCall: child stopped", "pid", m.pid)
			s.deregisterChild(m.pid)
			return server.Reply[gotp.Msg](nil), server.NoCont[any](), nil
		}
	}

	panic(fmt.Sprintf("StaticSupervisor.HandleCall: unknown message type %T", msg))
}

func (s *StaticSupervisor) HandleCast(msg gotp.Msg) (cont server.Continue[any], err error) {
	return server.NoCont[any](), nil
}

func (s *StaticSupervisor) HandleContinue(arg any) (cont server.Continue[any], err error) {
	return server.NoCont[any](), nil
}

func (s *StaticSupervisor) HandleInfo(info gotp.Msg) (cont server.Continue[any], err error) {
	switch info := info.(type) {
	case gotp.Exit:
		if child, ok := s.findChildByPID(info.PID()); !ok {
			slog.Debug("StaticSupervisor.HandleInfo: exit from unknown child", "pid", info.PID())
			return server.NoCont[any](), nil
		} else if !s.deregisterChild(info.PID()) {
			slog.Debug("StaticSupervisor.HandleInfo: exit child not found", "pid", info.PID())
			return server.NoCont[any](), nil
		} else if !s.shouldRestart(info.PID()) {
			slog.Debug("StaticSupervisor.HandleInfo: child should not be restarted", "pid", info.PID())
			return server.NoCont[any](), nil
		} else if err := s.StartChild(child.supervisable, 0); err != nil {
			slog.Error("StaticSupervisor.HandleInfo: failed to restart child", "pid", info.PID(), "error", err)
			// TODO: track this timer somewhere?
			_, err = s.server.SendAfter(server.Call[any, any](startChild{child.supervisable}, s.PID()), s.flags.ResetPeriod)
			return server.NoCont[any](), err
		}
		slog.Debug("StaticSupervisor.HandleInfo: child restarted", "pid", info.PID())
		return server.NoCont[any](), nil

	default:
		slog.Debug("StaticSupervisor.HandleInfo: received info message", "msg", info)
		return server.NoCont[any](), nil
	}
}

func (s *StaticSupervisor) HandleAny(msg gotp.Msg) (cont server.Continue[any], err error) {
	return server.NoCont[any](), nil
}

func (s *StaticSupervisor) Terminate(reason error) (newReson error) {
	slog.Debug("StaticSupervisor.Terminate: terminating supervisor", "reason", reason)
	timeout := time.After(s.flags.Shutdown)
shutdown:
	for pid, child := range s.children {
		child.running.Send(gotp.NewExit(pid, reason), 0)
		slog.Debug("StaticSupervisor.Terminate: sent exit to child", "pid", pid, "reason", reason)

		select {
		case msg := <-s.server.Receive():
			if msg, ok := msg.(gotp.Exit); ok {
				slog.Debug("StaticSupervisor.Terminate: received exit", "msg", msg)
				s.deregisterChild(msg.PID())
				continue
			}
			slog.Debug("StaticSupervisor.Terminate: received message", "msg", msg)
		case <-timeout:
			slog.Debug("StaticSupervisor.Terminate: timeout waiting for children to exit")
			break shutdown
		}
	}
	return reason
}

func (s *StaticSupervisor) findChild(child gotp.Supervisable) (gotp.PID, bool) {
	if child == nil {
		return gotp.PIDZero(), false
	}
	if pid, ok := s.childNames[child.ChildSpec().Name]; ok {
		return pid, true
	}
	return gotp.PIDZero(), false
}

func (s *StaticSupervisor) findChildByPID(pid gotp.PID) (child child, ok bool) {
	if pid == gotp.PIDZero() {
		return child, false
	}
	child, ok = s.children[pid]
	return child, ok
}

func (s *StaticSupervisor) startChild(child gotp.Supervisable, timeout time.Duration) (_ gotp.Running, reason error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("StaticSupervisor.startChild: panic", "error", r)
			if reason == nil {
				reason = fmt.Errorf("panic: %v", r)
			} else {
				reason = fmt.Errorf("reason: %w, panic: %v", reason, r)
			}
		}
	}()
	spec := child.ChildSpec()
	return child.StartLink(s.server.PID(), timeout, spec.SpawnOpts...)
}

func (s *StaticSupervisor) registerChild(supervisable gotp.Supervisable, running gotp.Running) {
	pid := running.PID()
	s.children[pid] = child{supervisable: supervisable, running: running}
	if name := supervisable.ChildSpec().Name; name != "" {
		s.childNames[name] = pid
	}
}

func (s *StaticSupervisor) deregisterChild(pid gotp.PID) (deleted bool) {
	if child, ok := s.children[pid]; ok {
		if name := child.supervisable.ChildSpec().Name; name != "" {
			delete(s.childNames, name)
		}
		delete(s.children, pid)
		deleted = true
		slog.Debug("StaticSupervisor.deregisterChild: child deregistered", "pid", pid)
	}
	return deleted
}

func (s *StaticSupervisor) shouldRestart(pid gotp.PID) (restart bool) {
	if s.server.PID() == pid {
		return false
	}

	cs, ok := s.children[pid]
	if !ok {
		return true
	}

	cs.restart.count++

	if cs.restart.count == 1 {
		cs.restart.at = time.Now()
		restart = true
	} else if cs.restart.count <= s.flags.MaxRestarts {
		restart = true
	} else if time.Since(cs.restart.at) > s.flags.ResetPeriod {
		cs.restart.count = 1
		cs.restart.at = time.Now()
		restart = true
	} else {
		restart = false
	}

	if restart {
		s.children[pid] = cs
	}

	return restart
}
