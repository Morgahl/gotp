package supervisor

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/process"
	"github.com/Morgahl/gotp/server"
)

var _ server.Supervisable = &StaticSupervisor{}
var _ server.Supervised = &StaticSupervisor{}
var _ server.Serverable[gotp.Options, process.Message, process.Message, process.Message, process.Message] = &StaticSupervisor{}

type StaticSupervisor struct {
	id       gotp.Atom
	sup      Supervisor[gotp.Options]
	flags    Flags
	specs    []server.Supervisable
	childIDs map[gotp.Atom]process.PID
	children map[process.PID]child
	server   *server.Server[gotp.Options, process.Message, process.Message, process.Message, process.Message]
}

func Static(id gotp.Atom, sup Supervisor[gotp.Options], initArg gotp.Options) *StaticSupervisor {
	static := &StaticSupervisor{id: id, sup: sup}
	static.server = server.New(static, initArg)
	return static
}

func (s *StaticSupervisor) Context() context.Context {
	return s.server.Process().Context()
}

func (s *StaticSupervisor) ChildSpec() server.ChildSpec {
	return s.sup.ChildSpec()
}

func (s *StaticSupervisor) Start(opts ...process.SpawnOpt) (process.Started, error) {
	return s.server.Start(opts...)
}

func (s *StaticSupervisor) StartLink(link *process.Process, opts ...process.SpawnOpt) (server.Supervised, error) {
	return s.server.StartLink(link, opts...)
}

func (s *StaticSupervisor) PID() process.PID {
	return s.server.PID()
}

func (s *StaticSupervisor) Send(msg process.Message) {
	s.server.Send(msg)
}

func (s *StaticSupervisor) SendAfter(msg process.Message, delay time.Duration) *time.Timer {
	return s.server.SendAfter(msg, delay)
}

func (s *StaticSupervisor) Exit(reason error) {
	s.server.Exit(reason)
}

func (s *StaticSupervisor) StartChild(child server.Supervisable) process.Message {
	self := s.PID()
	if resp, ok := server.Call[process.Message, process.Message](self, self, startChild{child}, 0); ok {
		switch resp := resp.(type) {
		case process.PID, server.AlreadyStarted, error:
			return resp
		default:
			panic(fmt.Sprintf("StaticSupervisor.StartChild: unexpected response type %T", resp))
		}
	}
	return nil
}

func (s *StaticSupervisor) StopChild(pid process.PID) process.Message {
	self := s.PID()
	if resp, ok := server.Call[process.Message, process.Message](self, self, stopChild{pid}, 0); ok {
		switch resp := resp.(type) {
		case bool:
			return resp
		default:
			panic(fmt.Sprintf("StaticSupervisor.StopChild: unexpected response type %T", resp))
		}
	}
	return nil
}

func (s *StaticSupervisor) Init(opts gotp.Options) (cont server.Continue[process.Message], err error) {
	s.server.Process().UpdateFlags(func(flags process.ProcessFlags) process.ProcessFlags {
		return flags | process.TRAP_EXIT_FLAG
	})

	var flags Flags
	var children []server.Supervisable

	if flags, children, err = s.sup.Init(opts); err != nil {
		return server.NoCont[process.Message](), err
	}
	s.flags = flags.ApplyDefaults()
	s.specs = children
	s.childIDs = make(map[gotp.Atom]process.PID, len(children))
	s.children = make(map[process.PID]child, len(children))

	for _, child := range s.specs {
		if child == nil {
			continue
		} else if supervised, err := s.startChild(child); err != nil {
			return server.NoCont[process.Message](), err
		} else {
			s.registerChild(supervised)
		}
	}

	return server.NoCont[process.Message](), nil
}

func (s *StaticSupervisor) HandleCall(msg process.Message, _ process.PID) (resp server.Response[process.Message], cont server.Continue[process.Message], err error) {
	switch m := msg.(type) {
	case startChild:
		if pid, ok := s.findChild(m.child); ok {
			return server.Reply[process.Message](server.NewAlreadyStarted(pid)), server.NoCont[process.Message](), nil
		} else if supervised, err := s.startChild(m.child); err != nil {
			return server.Reply[process.Message](err), server.NoCont[process.Message](), nil
		} else {
			s.registerChild(supervised)
			return server.Reply[process.Message](supervised.PID()), server.NoCont[process.Message](), nil
		}

	case stopChild:
		child, ok := s.findChildByPID(m.pid)
		if !ok {
			return server.Reply[process.Message](false), server.NoCont[process.Message](), nil
		}
		child.supervised.Send(process.ExitMsg{PID: s.PID(), Reason: process.KILL})
		removed := s.deregisterChild(m.pid)
		return server.Reply[process.Message](removed), server.NoCont[process.Message](), nil
	}

	panic(fmt.Sprintf("StaticSupervisor.HandleCall: unknown message type %T", msg))
}

func (s *StaticSupervisor) HandleCast(msg process.Message) (cont server.Continue[process.Message], err error) {
	return server.NoCont[process.Message](), nil
}

func (s *StaticSupervisor) HandleContinue(arg process.Message) (cont server.Continue[process.Message], err error) {
	return server.NoCont[process.Message](), nil
}

func (s *StaticSupervisor) HandleInfo(info process.Message) (cont server.Continue[process.Message], err error) {
	defer func() {
		if r := recover(); r != nil {
		}
	}()
	switch info := info.(type) {
	case process.ExitMsg:
		if info.PID == s.server.PID() {
			return server.NoCont[process.Message](), info.Reason
		}

		child, ok := s.findChildByPID(info.PID)
		if !ok {
			return server.NoCont[process.Message](), nil
		}
		// need to call this here as the deregisterChild will remove the child from the map
		shouldRestart := s.shouldRestart(info.PID, info.Reason)
		if !s.deregisterChild(info.PID) {
			return server.NoCont[process.Message](), nil
		} else if !shouldRestart {
			return server.NoCont[process.Message](), nil
		}
		switch s.StartChild(child.supervised).(type) {
		case process.PID, server.AlreadyStarted:
			return server.NoCont[process.Message](), nil
		case error:
			// TODO: track this timer somewhere?
			// TODO: also this likely need to be a computed reset for the after timer
			_ = s.server.SendAfter(server.CallMsg[process.Message, process.Message](s.PID(), startChild{child.supervised}), s.flags.ResetPeriod)
			return server.NoCont[process.Message](), err

		default:
			return server.NoCont[process.Message](), nil
		}
	}
	return server.NoCont[process.Message](), nil
}

func (s *StaticSupervisor) Terminate(reason error) (newReson error) {
	children := make([]child, 0, len(s.children))
	for _, c := range s.children {
		children = append(children, c)
	}

	// TODO: ultimately we want to maintin the processes as a stack and reap them in reverse order
	// TODO: for now, we sort them by PID in descending order to ensure that the most recently started
	// TODO: processes are terminated first
	// TODO: we may be able to borrow lazy GC practices from the process.Process's mailbox cleanup to
	// TODO: avoid the need to sort the children here as well as keep an ordered stack of children
	// TODO: for the correct one_for_one, one_for_all, and rest_for_one handling
	slices.SortFunc(children, func(i, j child) int {
		return process.ComparePID(i.supervised.PID(), j.supervised.PID()) * -1
	})

	for _, child := range children {
		pid := child.supervised.PID()
		timeout := child.supervised.ChildSpec().Shutdown
		child.supervised.Send(process.ExitMsg{PID: pid, Reason: reason})

		if msg, ok, err := process.ReceiveWithTimeout[process.ExitMsg](s.server.Process(), timeout); err != nil {
			return err
		} else if ok {
			s.deregisterChild(msg.PID)
			continue
		} else {
			s.deregisterChild(msg.PID)
		}
	}
	return reason
}

func (s *StaticSupervisor) findChild(child server.Supervisable) (process.PID, bool) {
	if child == nil {
		return process.PIDZero(), false
	}
	if pid, ok := s.childIDs[child.ChildSpec().ID]; ok {
		return pid, true
	}
	return process.PIDZero(), false
}

func (s *StaticSupervisor) findChildByPID(pid process.PID) (child child, ok bool) {
	if pid == process.PIDZero() {
		return child, false
	}
	child, ok = s.children[pid]
	return child, ok
}

func (s *StaticSupervisor) startChild(child server.Supervisable) (server.Supervised, error) {
	return child.StartLink(s.server.Process(), child.ChildSpec().SpawnOpts...)
}

func (s *StaticSupervisor) registerChild(supervised server.Supervised) {
	pid := supervised.PID()
	s.children[pid] = child{supervised: supervised}
	if id := supervised.ChildSpec().ID; id != "" {
		s.childIDs[id] = pid
	}
}

func (s *StaticSupervisor) deregisterChild(pid process.PID) (deleted bool) {
	if child, ok := s.children[pid]; ok {
		if id := child.supervised.ChildSpec().ID; id != "" {
			delete(s.childIDs, id)
		}
		delete(s.children, pid)
		deleted = true
	}
	return deleted
}

func (s *StaticSupervisor) shouldRestart(pid process.PID, reason error) (restart bool) {
	if s.server.PID() == pid {
		return false
	}

	cs, ok := s.children[pid]
	if !ok {
		return false
	}
	spec := cs.supervised.ChildSpec()
	switch spec.Restart {
	case server.TEMPORARY, server.TRANSIENT:
		return false
	case server.PERMANENT:
		if errors.Is(reason, process.KILL) || errors.Is(reason, process.NORMAL) && spec.Type == server.WORKER {
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
	supervised server.Supervised
	restart    struct {
		count uint64
		at    time.Time
	}
}
