package supervisor

import (
	"fmt"
	"slices"
	"time"

	gotp "github.com/Morgahl/gotp/old"
	"github.com/Morgahl/gotp/old/server"
	"github.com/Morgahl/gotp/process"
)

var _ gotp.Supervisable = &StaticSupervisor{}
var _ gotp.Supervised = &StaticSupervisor{}
var _ server.Serverable[gotp.Options, gotp.Msg, gotp.Msg, gotp.Msg, gotp.Msg] = &StaticSupervisor{}

type StaticSupervisor struct {
	id       gotp.Atom
	sup      Supervisor[gotp.Options]
	flags    Flags
	specs    []gotp.Supervisable
	childIDs map[gotp.Atom]gotp.PID
	children map[gotp.PID]child
	server   *server.Server[gotp.Options, gotp.Msg, gotp.Msg, gotp.Msg, gotp.Msg]
}

func Static(id gotp.Atom, sup Supervisor[gotp.Options], initArg gotp.Options) *StaticSupervisor {
	static := &StaticSupervisor{id: id, sup: sup}
	static.server = server.New(static, initArg)
	return static
}

func (s *StaticSupervisor) ChildSpec() gotp.ChildSpec {
	return gotp.ChildSpec{
		ID:          s.id,
		Restart:     gotp.PERMANENT,
		Shutdown:    gotp.DEFAULT_SHUTDOWN,
		Type:        gotp.SUPERVISOR,
		Significant: true,
	}
}

func (s *StaticSupervisor) Start(opts ...gotp.SpawnOpt) (gotp.Started, error) {
	return s.server.Start(opts...)
}

func (s *StaticSupervisor) StartLink(link gotp.PID, opts ...gotp.SpawnOpt) (gotp.Supervised, error) {
	return s.server.StartLink(link, opts...)
}

func (s *StaticSupervisor) ID() gotp.PID {
	return s.server.ID()
}

func (s *StaticSupervisor) Send(msg gotp.Msg) {
	s.server.Send(msg)
}

func (s *StaticSupervisor) SendAfter(msg gotp.Msg, delay time.Duration) *time.Timer {
	return s.server.SendAfter(msg, delay)
}

func (s *StaticSupervisor) Receive() <-chan gotp.Msg {
	return s.server.Receive()
}

func (s *StaticSupervisor) StartChild(child gotp.Supervisable) gotp.Msg {
	self := s.ID()
	if resp, ok := server.Call[gotp.Msg, gotp.Msg](self, self, startChild{child}); ok {
		switch resp := resp.(type) {
		case gotp.PID, gotp.AlreadyStarted, error:
			return resp
		default:
			panic(fmt.Sprintf("StaticSupervisor.StartChild: unexpected response type %T", resp))
		}
	}
	return nil
}

func (s *StaticSupervisor) StopChild(pid gotp.PID) gotp.Msg {
	self := s.ID()
	if resp, ok := server.Call[gotp.Msg, gotp.Msg](self, self, stopChild{pid}); ok {
		switch resp := resp.(type) {
		case bool:
			return resp
		default:
			panic(fmt.Sprintf("StaticSupervisor.StopChild: unexpected response type %T", resp))
		}
	}
	return nil
}

func (s *StaticSupervisor) Init(opts gotp.Options) (cont server.Continue[gotp.Msg], err error) {
	var flags Flags
	var children []gotp.Supervisable

	if flags, children, err = s.sup.Init(opts); err != nil {
		return server.NoCont[gotp.Msg](), err
	}
	s.flags = flags.ApplyDefaults()
	s.specs = children
	s.childIDs = make(map[gotp.Atom]gotp.PID, len(children))
	s.children = make(map[gotp.PID]child, len(children))

	for _, child := range s.specs {
		if child == nil {
			continue
		} else if supervised, err := s.startChild(child); err != nil {
			return server.NoCont[gotp.Msg](), err
		} else {
			s.registerChild(supervised)
		}
	}

	return server.NoCont[gotp.Msg](), nil
}

func (s *StaticSupervisor) HandleCall(msg gotp.Msg, _ gotp.PID) (resp server.Response[gotp.Msg], cont server.Continue[gotp.Msg], err error) {
	switch m := msg.(type) {
	case startChild:
		if pid, ok := s.findChild(m.child); ok {
			return server.Reply[gotp.Msg](gotp.NewAlreadyStarted(pid)), server.NoCont[gotp.Msg](), nil
		} else if supervised, err := s.startChild(m.child); err != nil {
			return server.Reply[gotp.Msg](err), server.NoCont[gotp.Msg](), nil
		} else {
			s.registerChild(supervised)
			return server.Reply[gotp.Msg](supervised.ID()), server.NoCont[gotp.Msg](), nil
		}

	case stopChild:
		child, ok := s.findChildByPID(m.pid)
		if !ok {
			return server.Reply[gotp.Msg](false), server.NoCont[gotp.Msg](), nil
		}
		child.supervised.Send(gotp.NewExit(m.pid, gotp.Kill{}))
		removed := s.deregisterChild(m.pid)
		return server.Reply[gotp.Msg](removed), server.NoCont[gotp.Msg](), nil
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
		child, ok := s.findChildByPID(info.ID())
		if !ok {
			return server.NoCont[gotp.Msg](), nil
		}
		shouldRestart := s.shouldRestart(info.ID(), info.Reason())
		if !s.deregisterChild(info.ID()) {
			return server.NoCont[gotp.Msg](), nil
		} else if !shouldRestart {
			return server.NoCont[gotp.Msg](), nil
		}
		r := s.StartChild(child.supervised)
		switch r.(type) {
		case gotp.PID, gotp.AlreadyStarted:
			return server.NoCont[gotp.Msg](), nil
		case error:
			// TODO: track this timer somewhere?
			_ = s.server.SendAfter(server.CallMsg[gotp.Msg, gotp.Msg](s.ID(), startChild{child.supervised}), s.flags.ResetPeriod)
			return server.NoCont[gotp.Msg](), err

		default:
			return server.NoCont[gotp.Msg](), nil
		}
	}
	return server.NoCont[gotp.Msg](), nil
}

func (s *StaticSupervisor) Terminate(reason error) (newReson error) {
	children := make([]child, 0, len(s.children))
	for _, c := range s.children {
		children = append(children, c)
	}

	// TODO: ultimately we want to maintin the processes as a stack and reap them in reverse order
	// TODO: for now, we sort them by PID in descending order to ensure that the most recently started
	// TODO: processes are terminated first
	slices.SortFunc(children, func(i, j child) int {
		return gotp.ComparePID(i.supervised.ID(), j.supervised.ID()) * -1
	})

	for _, child := range children {
		pid := child.supervised.ID()
		timeout := child.supervised.ChildSpec().Shutdown
		child.supervised.Send(gotp.NewExit(pid, reason))

		select {
		case msg, ok := <-s.server.Receive():
			if !ok {
				panic("StaticSupervisor.Terminate: mailbox closed unexpectedly")
			}
			if msg, ok := msg.(gotp.Exit); ok {
				s.deregisterChild(msg.ID())
				continue
			}

		case <-time.After(timeout):
			continue
		}
	}
	return reason
}

func (s *StaticSupervisor) findChild(child gotp.Supervisable) (gotp.PID, bool) {
	if child == nil {
		return gotp.PIDZero(), false
	}
	if pid, ok := s.childIDs[child.ChildSpec().ID]; ok {
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

func (s *StaticSupervisor) startChild(child gotp.Supervisable) (gotp.Supervised, error) {
	return child.StartLink(s.server.ID(), child.ChildSpec().SpawnOpts...)
}

func (s *StaticSupervisor) registerChild(supervised gotp.Supervised) {
	pid := supervised.ID()
	s.children[pid] = child{supervised: supervised}
	if id := supervised.ChildSpec().ID; id != "" {
		s.childIDs[id] = pid
	}
}

func (s *StaticSupervisor) deregisterChild(pid gotp.PID) (deleted bool) {
	if child, ok := s.children[pid]; ok {
		if id := child.supervised.ChildSpec().ID; id != "" {
			delete(s.childIDs, id)
		}
		delete(s.children, pid)
		deleted = true
	}
	return deleted
}

func (s *StaticSupervisor) shouldRestart(pid gotp.PID, reason error) (restart bool) {
	if s.server.ID() == pid {
		return false
	}

	cs, ok := s.children[pid]
	if !ok {
		return false
	}

	switch cs.supervised.ChildSpec().Restart {
	case gotp.TEMPORARY, gotp.TRANSIENT:
		return false
	case gotp.PERMANENT:
		if reason == process.NORMAL {
			return false
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
	}

	return restart
}

type child struct {
	supervised gotp.Supervised
	restart    struct {
		count uint
		at    time.Time
	}
}
