package supervisor

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/server"
)

var _ gotp.Supervisable = &StaticSupervisor{}
var _ gotp.Supervised = &StaticSupervisor{}
var _ server.Serverable[gotp.Msg, gotp.Msg, gotp.Msg, gotp.Msg, gotp.Msg] = &StaticSupervisor{}

type StaticSupervisor struct {
	sup        Supervisor
	flags      Flags
	specs      []gotp.Supervisable
	childNames map[string]gotp.PID
	children   map[gotp.PID]child
	server     *server.Server[gotp.Msg, gotp.Msg, gotp.Msg, gotp.Msg, gotp.Msg]
}

func Static(sup Supervisor) *StaticSupervisor {
	static := &StaticSupervisor{sup: sup}
	static.server = server.New(static)
	return static
}

func (s *StaticSupervisor) ChildSpec() gotp.ChildSpec {
	return gotp.ChildSpec{
		Restart:     gotp.PERMANENT,
		Shutdown:    30 * time.Second,
		Type:        gotp.SUPERVISOR,
		Significant: true,
	}
}

func (s *StaticSupervisor) Start(timeout time.Duration, opts ...gotp.SpawnOpt) (gotp.Started, error) {
	return s.server.Start(timeout, opts...)
}

func (s *StaticSupervisor) StartLink(link gotp.PID, timeout time.Duration, opts ...gotp.SpawnOpt) (gotp.Supervised, error) {
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

func (s *StaticSupervisor) Init(opts gotp.Options) (cont server.Continue[gotp.Msg], err error) {
	var flags Flags
	var children []gotp.Supervisable
	if flags, children, err = s.sup.Init(opts); err != nil {
		slog.Error("StaticSupervisor.Init: failed to initialize supervisor", "error", err)
		return server.NoCont[gotp.Msg](), err
	}
	s.flags = flags.ApplyDefaults()
	s.specs = children
	s.childNames = make(map[string]gotp.PID, len(children))
	s.children = make(map[gotp.PID]child, len(children))

	for _, child := range s.specs {
		if child == nil {
			slog.Error("StaticSupervisor.Init: child is nil, skipping")
			continue
		} else if running, err := s.startChild(child, 0); err != nil {
			slog.Error("StaticSupervisor.Init: failed to start child", "error", err)
			return server.NoCont[gotp.Msg](), err
		} else {
			s.registerChild(child, running)
		}
	}
	return server.NoCont[gotp.Msg](), nil
}

func (s *StaticSupervisor) HandleCall(msg gotp.Msg, _ gotp.PID) (resp server.Response[gotp.Msg], cont server.Continue[gotp.Msg], err error) {
	switch m := msg.(type) {
	case startChild:
		if pid, ok := s.findChild(m.child); ok {
			return server.Reply[gotp.Msg](gotp.NewAlreadyStarted(pid)), server.NoCont[gotp.Msg](), nil
		} else if running, err := s.startChild(m.child, 0); err != nil {
			slog.Error("StaticSupervisor.HandleCall: failed to start child", "error", err)
			return server.Reply[gotp.Msg](err), server.NoCont[gotp.Msg](), nil
		} else {
			s.registerChild(m.child, running)
			return server.Reply[gotp.Msg](running.PID()), server.NoCont[gotp.Msg](), nil
		}

	case stopChild:
		if child, ok := s.findChildByPID(m.pid); !ok {
			return server.Reply[gotp.Msg](false), server.NoCont[gotp.Msg](), nil
		} else if err := child.running.Exit(gotp.Kill{}, 0); err != nil {
			slog.Error("StaticSupervisor.HandleCall: failed to stop child", "pid", m.pid, "error", err)
			return server.Reply[gotp.Msg](err), server.NoCont[gotp.Msg](), nil
		} else {
			s.deregisterChild(m.pid)
			return server.Reply[gotp.Msg](nil), server.NoCont[gotp.Msg](), nil
		}
	}

	panic(fmt.Sprintf("StaticSupervisor.HandleCall: unknown message type %T", msg))
}

func (s *StaticSupervisor) HandleCast(msg gotp.Msg) (cont server.Continue[gotp.Msg], err error) {
	return server.NoCont[gotp.Msg](), nil
}

func (s *StaticSupervisor) HandleContinue(arg gotp.Msg) (cont server.Continue[gotp.Msg], err error) {
	return server.NoCont[gotp.Msg](), nil
}

func (s *StaticSupervisor) HandleInfo(info gotp.Msg) (cont server.Continue[gotp.Msg], err error) {
	switch info := info.(type) {
	case gotp.Exit:
		if child, ok := s.findChildByPID(info.PID()); !ok {
			return server.NoCont[gotp.Msg](), nil
		} else if !s.deregisterChild(info.PID()) {
			return server.NoCont[gotp.Msg](), nil
		} else if !s.shouldRestart(info.PID()) {
			return server.NoCont[gotp.Msg](), nil
		} else if err := s.StartChild(child.supervisable, 0); err != nil {
			slog.Error("StaticSupervisor.HandleInfo: failed to restart child", "pid", info.PID(), "error", err)
			// TODO: track this timer somewhere?
			_, err = s.server.SendAfter(server.Call[any, any](startChild{child.supervisable}, s.PID()), s.flags.ResetPeriod)
			return server.NoCont[gotp.Msg](), err
		}
		return server.NoCont[gotp.Msg](), nil

	default:
		return server.NoCont[gotp.Msg](), nil
	}
}

func (s *StaticSupervisor) HandleAny(msg gotp.Msg) (cont server.Continue[gotp.Msg], err error) {
	return server.NoCont[gotp.Msg](), nil
}

func (s *StaticSupervisor) Terminate(reason error) (newReson error) {
	timeout := time.After(s.flags.Shutdown)
shutdown:
	for pid, child := range s.children {
		child.running.Send(gotp.NewExit(pid, reason), 0)

		select {
		case msg := <-s.server.Receive():
			if msg, ok := msg.(gotp.Exit); ok {
				s.deregisterChild(msg.PID())
				continue
			}
		case <-timeout:
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

func (s *StaticSupervisor) startChild(child gotp.Supervisable, timeout time.Duration) (_ gotp.Supervised, reason error) {
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

func (s *StaticSupervisor) registerChild(supervisable gotp.Supervisable, running gotp.Started) {
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
