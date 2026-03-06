package static

import (
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/dbg"
	"github.com/Morgahl/gotp/gen_server"
	"github.com/Morgahl/gotp/internal/ctx"
	"github.com/Morgahl/gotp/process"
	"github.com/Morgahl/gotp/supervisor"
	"github.com/Morgahl/gotp/term"
)

var _ gen_server.GenServer[any, term.Term, term.Term, term.Term, term.Term] = &server[any]{}

type server[I any] struct {
	sup      supervisor.Supervisor[I]
	options  supervisor.Options
	specs    []supervisor.ChildSpec
	childIDs map[gotp.Atom]process.PID
	children map[process.PID]child
	sigCount uint64
}

func (s *server[I]) ChildSpec() supervisor.ChildSpec {
	return s.sup.ChildSpec()
}

func (s *server[I]) Init(pctx process.Context, opts I) (cont gen_server.Continue[term.Term], err error) {
	pctx.TrapExit(true)

	var options supervisor.Options
	var children []supervisor.Supervisable
	if options, children, err = s.sup.Init(pctx, opts); err != nil {
		return gen_server.NoCont[term.Term](), err
	}

	s.options = options.ApplyDefaults()
	s.specs = make([]supervisor.ChildSpec, 0, len(children))
	s.childIDs = make(map[gotp.Atom]process.PID, len(children))
	s.children = make(map[process.PID]child, len(children))
	for _, child := range children {
		spec := child.ChildSpec()
		s.specs = append(s.specs, spec)
		if err := s.startChild(pctx, spec); err != nil {
			return gen_server.NoCont[term.Term](), err
		}
		pctx.ProcessPending()
	}

	return gen_server.NoCont[term.Term](), nil
}

func (s *server[I]) HandleCall(pctx process.Context, msg term.Term, _ process.PID) (resp gen_server.Response[term.Term], cont gen_server.Continue[term.Term], err error) {
	switch m := msg.(type) {
	// starting child
	case supervisor.ChildSpec:
		if pid, ok := s.findChild(m); ok {
			return gen_server.Reply[term.Term](supervisor.NewAlreadyStarted(pid)), gen_server.NoCont[term.Term](), nil
		} else if err := s.startChild(pctx, m); err != nil {
			return gen_server.Reply[term.Term](err), gen_server.NoCont[term.Term](), nil
		} else if pid, ok := s.findChild(m); ok {
			return gen_server.Reply[term.Term](pid), gen_server.NoCont[term.Term](), nil
		} else {
			return gen_server.Reply[term.Term](supervisor.NewNotStarted()), gen_server.NoCont[term.Term](), nil
		}

	// stopping child
	case process.PID:
		if child, ok := s.findChildByPID(m); !ok {
			return gen_server.Reply[term.Term](false), gen_server.NoCont[term.Term](), nil
		} else {
			child.ref.Send(process.ExitMsg{PID: m, Reason: process.KILL})
			removed := s.deregisterChild(m)
			return gen_server.Reply[term.Term](removed), gen_server.NoCont[term.Term](), nil
		}
	}

	dbg.Throw("StaticSupervisor.HandleCall: unknown message type %T", msg)
	panic("unreachable code")
}

func (s *server[I]) HandleCast(pctx process.Context, msg term.Term) (cont gen_server.Continue[term.Term], err error) {
	return gen_server.NoCont[term.Term](), nil
}

func (s *server[I]) HandleContinue(pctx process.Context, arg term.Term) (cont gen_server.Continue[term.Term], err error) {
	return gen_server.NoCont[term.Term](), nil
}

func (s *server[I]) HandleInfo(pctx process.Context, info term.Term) (cont gen_server.Continue[term.Term], err error) {
	switch info := info.(type) {
	case process.ExitMsg:
		if info.PID == pctx.PID() {
			return gen_server.NoCont[term.Term](), info.Reason
		}

		child, ok := s.findChildByPID(info.PID)
		if !ok {
			return gen_server.NoCont[term.Term](), nil
		}
		// need to call this here as the deregisterChild will remove the child from the map
		shouldRestart := s.shouldRestart(info.PID, info.Reason)
		if !s.deregisterChild(info.PID) {
			return gen_server.NoCont[term.Term](), nil
		} else if !shouldRestart {
			if s.options.AutoShutdown == supervisor.ALL_SIGNIFICANT && s.sigCount == 0 {
				slog.DebugContext(pctx.Context(), "StaticSupervisor.HandleInfo - no children left, stopping")
				return gen_server.Stop[term.Term](process.NORMAL), nil
			}
			return gen_server.NoCont[term.Term](), nil
		}

		switch s.options.Strategy {
		case supervisor.ONE_FOR_ONE:
			return s.restartOneForOne(pctx, child)
		case supervisor.ONE_FOR_ALL:
			return s.restartOneForAll(pctx, child)
		case supervisor.REST_FOR_ONE:
			return s.restartRestForOne(pctx, child)
		}
	}

	return gen_server.NoCont[term.Term](), nil
}

func (s *server[I]) restartOneForOne(pctx process.Context, failed child) (gen_server.Continue[term.Term], error) {
	if err := s.startChild(pctx, failed.spec); err != nil {
		_ = pctx.SendAfter(gen_server.CallMsg[term.Term, term.Term](pctx.PID(), failed), s.options.ResetPeriod)
		return gen_server.NoCont[term.Term](), err
	}
	return gen_server.NoCont[term.Term](), nil
}

func (s *server[I]) restartOneForAll(pctx process.Context, failed child) (gen_server.Continue[term.Term], error) {
	s.stopAllChildren(pctx)
	for _, spec := range s.specs {
		if err := s.startChild(pctx, spec); err != nil {
			return gen_server.NoCont[term.Term](), err
		}
		pctx.ProcessPending()
	}
	return gen_server.NoCont[term.Term](), nil
}

func (s *server[I]) restartRestForOne(pctx process.Context, failed child) (gen_server.Continue[term.Term], error) {
	// Find the position of the failed child's spec in the ordered spec list.
	failedIdx := -1
	for i, spec := range s.specs {
		if spec.Equals(failed.spec) {
			failedIdx = i
			break
		}
	}
	if failedIdx == -1 {
		// Spec not in the ordered list; fall back to one-for-one.
		return s.restartOneForOne(pctx, failed)
	}

	// Stop children whose specs come after the failed child (reverse order).
	specsToRestart := s.specs[failedIdx:]
	for i := len(specsToRestart) - 1; i >= 0; i-- {
		spec := specsToRestart[i]
		if pid, ok := s.childIDs[spec.ID]; ok {
			s.stopChild(pctx, pid)
		}
	}

	// Restart the failed child and all subsequent children in spec order.
	for _, spec := range specsToRestart {
		if err := s.startChild(pctx, spec); err != nil {
			return gen_server.NoCont[term.Term](), err
		}
		pctx.ProcessPending()
	}
	return gen_server.NoCont[term.Term](), nil
}

func (s *server[I]) stopAllChildren(pctx process.Context) {
	children := make([]child, 0, len(s.children))
	for _, c := range s.children {
		children = append(children, c)
	}
	slices.SortFunc(children, func(i, j child) int {
		return process.ComparePID(i.ref.PID(), j.ref.PID()) * -1
	})
	for _, c := range children {
		s.stopChild(pctx, c.ref.PID())
	}
}

func (s *server[I]) stopChild(pctx process.Context, pid process.PID) {
	c, ok := s.findChildByPID(pid)
	if !ok {
		return
	}
	timeout := c.spec.Shutdown
	c.ref.Send(process.ExitMsg{PID: pid, Reason: process.KILL})
	if msg, ok, _ := process.ReceiveWithTimeout[process.ExitMsg](pctx, timeout); ok {
		s.deregisterChild(msg.PID)
	} else {
		s.deregisterChild(pid)
	}
}

func (s *server[I]) Terminate(pctx process.Context, reason error) (newReson error) {
	children := make([]child, 0, len(s.children))
	for _, c := range s.children {
		children = append(children, c)
	}

	// TODO: ultimately we want to maintain the processes as a stack and reap them in reverse order
	// TODO: for now, we sort them by PID in descending order to ensure that the most recently started
	// TODO: processes are terminated first
	// TODO: we may be able to borrow lazy GC practices from the process.Process's mailbox cleanup to
	// TODO: avoid the need to sort the children here as well as keep an ordered stack of children
	// TODO: for the correct one_for_one, one_for_all, and rest_for_one handling
	slices.SortFunc(children, func(i, j child) int {
		return process.ComparePID(i.ref.PID(), j.ref.PID()) * -1
	})

	for _, child := range children {
		pid := child.ref.PID()
		timeout := child.spec.Shutdown
		child.ref.Send(process.ExitMsg{PID: pid, Reason: reason})

		if msg, ok, err := process.ReceiveWithTimeout[process.ExitMsg](pctx, timeout); err != nil {
			return err
		} else if ok {
			s.deregisterChild(msg.PID)
			continue
		} else {
			s.deregisterChild(pid)
		}
	}
	return reason
}

func (s *server[I]) findChild(spec supervisor.ChildSpec) (process.PID, bool) {
	for _, child := range s.children {
		if child.spec.Equals(spec) {
			pid := child.ref.PID()
			return pid, !pid.IsZero()
		}
	}
	return process.PIDZero(), false
}

func (s *server[I]) findChildByPID(pid process.PID) (child child, ok bool) {
	if pid == process.PIDZero() {
		return child, false
	}
	child, ok = s.children[pid]
	return child, ok
}

func (s *server[I]) startChild(pctx process.Context, spec supervisor.ChildSpec) error {
	ref, err := spec.Start(process.Link(pctx.Ref()))
	if err != nil {
		return err
	}
	return s.registerChild(spec, ref)
}

func (s *server[I]) registerChild(spec supervisor.ChildSpec, ref process.Ref) error {
	pid := ref.PID()
	if _, found := s.children[pid]; found {
		// TODO: this needs to be a well defined error type
		return fmt.Errorf("child with PID %s already registered", pid)
	}
	s.children[pid] = child{spec: spec, ref: ref}
	if id := spec.ID; id != "" {
		s.childIDs[id] = pid
	}
	if spec.Significant {
		s.sigCount++
	}
	return nil
}

func (s *server[I]) deregisterChild(pid process.PID) (deleted bool) {
	if child, ok := s.children[pid]; ok {
		if child.spec.Significant {
			s.sigCount--
		}
		if id := child.spec.ID; id != "" {
			delete(s.childIDs, id)
		}
		delete(s.children, pid)
		deleted = true
	}
	return deleted
}

// shouldRestart checks if the child should be restarted based on the restart strategy; assumes that the PID is not the supervisor's PID
func (s *server[I]) shouldRestart(pid process.PID, reason error) (restart bool) {
	cs, ok := s.children[pid]
	if !ok {
		return false
	}
	if errors.Is(reason, process.NORMAL) || errors.Is(reason, ctx.Shutdown{}) {
		return false
	}

	spec := cs.spec
	switch spec.Restart {
	case supervisor.TEMPORARY, supervisor.TRANSIENT:
		return false
	case supervisor.PERMANENT:
		cs.restart.count++
		if cs.restart.count == 1 {
			cs.restart.at = time.Now()
			restart = true
		} else if cs.restart.count <= s.options.MaxRestarts {
			restart = true
		} else if time.Since(cs.restart.at) > s.options.ResetPeriod {
			cs.restart.count = 1
			cs.restart.at = time.Now()
			restart = true
		} else {
			restart = false
		}

		s.children[pid] = cs
	}

	return restart
}

type child struct {
	spec    supervisor.ChildSpec
	ref     process.Ref
	restart struct {
		count uint64
		at    time.Time
	}
}
