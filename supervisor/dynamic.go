package supervisor

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Morgahl/gotp"
)

type startChild struct {
	child Supervisable
}

type stopChild struct {
	pid gotp.PID
}

var _ Supervisor = &DynamicSupervisor{}

type DynamicSupervisor struct {
	flags Flags

	process  *gotp.Process
	children map[gotp.PID]child
}

func Dynamic(flags Flags) *DynamicSupervisor {
	return &DynamicSupervisor{
		flags:    flags.ApplyDefaults(),
		children: make(map[gotp.PID]child),
	}
}

func (s *DynamicSupervisor) StartLink(link gotp.PID, timeout time.Duration, opts ...gotp.SpawnOpt) (Supervisable, error) {
	slog.Debug("DynamicSupervisor.StartLink", "link", link.String())
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

func (s *DynamicSupervisor) ID() gotp.PID {
	return s.process.ID()
}

func (s *DynamicSupervisor) ChildSpec() ChildSpec {
	return ChildSpec{
		Restart:     PERMANENT,
		Shutdown:    30 * time.Second,
		Type:        SUPERVISOR,
		Significant: true,
	}
}

func (s *DynamicSupervisor) StartChild(child Supervisable, timeout time.Duration) error {
	return s.Send(startChild{child}, timeout)
}

func (s *DynamicSupervisor) StopChild(pid gotp.PID, timeout time.Duration) error {
	return s.Send(stopChild{pid}, timeout)
}

func (s *DynamicSupervisor) Exit(reason error, timeout time.Duration) error {
	slog.Error("DynamicSupervisor.Exit", "reason", reason)
	return s.process.Send(gotp.NewExit(s.ID(), reason), timeout)
}

func (s *DynamicSupervisor) Exited() bool {
	slog.Debug("DynamicSupervisor.Exited")
	return s.process.Exited()
}

func (s *DynamicSupervisor) Send(msg gotp.Msg, timeout time.Duration) error {
	if s.process == nil {
		return errors.New("DynamicSupervisor not started")
	}
	return s.process.Send(msg, timeout)
}

func (s *DynamicSupervisor) setupProc(link gotp.PID, timeout time.Duration, opts ...gotp.SpawnOpt) <-chan struct{} {
	sig := make(chan struct{})
	s.process = gotp.SpawnLink(link, s.loop(sig), timeout, opts...)
	return sig
}

func (s *DynamicSupervisor) startChild(child Supervisable, timeout time.Duration) (reason error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("DynamicSupervisor.startChild: panic", "error", r)
			if reason == nil {
				reason = fmt.Errorf("panic: %v", r)
			} else {
				reason = fmt.Errorf("reason: %w, panic: %v", reason, r)
			}
		}
	}()
	spec := child.ChildSpec()
	child, err := child.StartLink(s.process.ID(), timeout, spec.SpawnOpts...)
	if err != nil {
		return err
	}
	s.registerChild(child)
	return nil
}

func (s *DynamicSupervisor) stopChild(pid gotp.PID, timeout time.Duration) {
	if child, exists := s.children[pid]; exists {
		child.supervisable.Send(gotp.NewExit(pid, nil), timeout)
	}
}

func (s *DynamicSupervisor) registerChild(supervisable Supervisable) {
	s.children[supervisable.ID()] = child{supervisable: supervisable}
}

func (s *DynamicSupervisor) deregisterChild(pid gotp.PID) {
	delete(s.children, pid)
}

func (s *DynamicSupervisor) shouldRestart(pid gotp.PID) (restart bool) {
	if s.process.ID() == pid {
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

func (s *DynamicSupervisor) shouldExit(pid gotp.PID) (exit bool) {
	return s.process.ID() == pid
}

func (s *DynamicSupervisor) loop(sig chan struct{}) gotp.RunFn {
	return func(_ *gotp.Process, in <-chan gotp.Msg) (reason error) {
		defer func() {
			if r := recover(); r != nil {
				if reason == nil {
					reason = fmt.Errorf("panic: %v", r)
				} else {
					reason = fmt.Errorf("reason: %w, panic: %v", reason, r)
				}
			}
			if len(s.children) > 0 {
				slog.Debug("DynamicSupervisor.loop: sending exits to children")
				timeout := time.After(s.flags.Shutdown)
			shutdown:
				for pid := range s.children {
					gotp.Send(pid, gotp.NewExit(pid, reason), 0)

					select {
					case msg := <-in:
						slog.Debug("DynamicSupervisor.loop: received message", "msg", msg)
						if msg, ok := msg.(gotp.Exit); ok {
							if msg.PID() == s.process.ID() {
								continue
							}
							slog.Debug("DynamicSupervisor.loop: received exit", "msg", msg)
							s.deregisterChild(msg.PID())
						}

					case <-timeout:
						slog.Debug("DynamicSupervisor.loop: timeout waiting for children to exit")
						break shutdown
					}
				}
			}
			slog.Debug("DynamicSupervisor.loop: exiting")
			if err := s.process.Exit(reason, 0); err != nil {
				slog.Error("DynamicSupervisor.loop: failed to exit process", "error", err)
			}
			slog.Debug("DynamicSupervisor.loop: exited")
		}()

		close(sig)

		var hasFailed bool
		var tickerRunning bool
		ticker := time.NewTicker(100 * time.Millisecond)
		ticker.Stop()
		defer ticker.Stop()

		for {
			if hasFailed && !tickerRunning {
				tickerRunning = true
				ticker.Reset(100 * time.Millisecond)
			} else if !hasFailed && tickerRunning {
				tickerRunning = false
				ticker.Stop()
			}

			select {

			case msg, ok := <-in:
				if !ok {
					slog.Debug("DynamicSupervisor.loop: input channel closed")
					return nil
				}
				slog.Debug("DynamicSupervisor.loop: received message", "msg", msg)
				switch msg := msg.(type) {
				case gotp.Exit:
					s.deregisterChild(msg.PID())
					if s.shouldRestart(msg.PID()) {
						if err := s.startChild(s.children[msg.PID()].supervisable, 0); err != nil {
							slog.Error("DynamicSupervisor.loop: failed to restart child", "error", err)
							hasFailed = true
						}
					} else if s.shouldExit(msg.PID()) {
						slog.Debug("DynamicSupervisor.loop: being told to exit")
						return msg.Unwrap()
					}

				case startChild:
					if err := s.startChild(msg.child, 0); err != nil {
						slog.Error("DynamicSupervisor.loop: failed to start child", "error", err)
						hasFailed = true
					}

				case stopChild:
					s.stopChild(msg.pid, 0)
				}

			case <-ticker.C:
				slog.Debug("DynamicSupervisor.loop: checking children")
				if !hasFailed {
					continue
				}

				var newFailed bool
				for pid := range s.children {
					if _, exists := s.children[pid]; !exists {
						if s.shouldRestart(pid) {
							if err := s.startChild(s.children[pid].supervisable, 0); err != nil {
								slog.Error("DynamicSupervisor.loop: failed to restart child", "error", err)
								newFailed = true
							}
						}
					}
				}
				hasFailed = newFailed
			}
		}
	}
}
