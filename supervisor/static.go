package supervisor

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Morgahl/gotp"
)

var _ Supervisor = &StaticSupervisor{}

type StaticSupervisor struct {
	flags Flags
	specs []Supervisable

	process    *gotp.Process
	childNames map[string]gotp.PID
	children   map[gotp.PID]child
}

func Static(flags Flags, children ...Supervisable) *StaticSupervisor {
	return &StaticSupervisor{
		flags:      flags.ApplyDefaults(),
		specs:      children,
		childNames: make(map[string]gotp.PID, len(children)),
		children:   make(map[gotp.PID]child, len(children)),
	}
}

func (s *StaticSupervisor) StartLink(link gotp.PID, timeout time.Duration, opts ...gotp.SpawnOpt) (Supervisable, error) {
	slog.Debug("StaticSupervisor.StartLink", "link", link.String())
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

func (s *StaticSupervisor) ID() gotp.PID {
	return s.process.ID()
}

func (s *StaticSupervisor) ChildSpec() ChildSpec {
	return ChildSpec{
		Restart:     PERMANENT,
		Shutdown:    30 * time.Second,
		Type:        SUPERVISOR,
		Significant: true,
	}
}

func (s *StaticSupervisor) StartChild(child Supervisable, timeout time.Duration) error {
	return errors.New("StaticSupervisor does not support dynamic child addition")
}

func (s *StaticSupervisor) StopChild(pid gotp.PID, timeout time.Duration) error {
	return errors.New("StaticSupervisor does not support dynamic child removal")
}

func (s *StaticSupervisor) Exit(reason error, timeout time.Duration) error {
	slog.Error("StaticSupervisor.Exit", "reason", reason)
	return s.process.Send(gotp.NewExit(s.ID(), reason), timeout)
}

func (s *StaticSupervisor) Exited() bool {
	slog.Debug("StaticSupervisor.Exited")
	return s.process.Exited()
}

func (s *StaticSupervisor) Send(msg gotp.Msg, timeout time.Duration) error {
	slog.Debug("StaticSupervisor.Send", "msg", msg)
	if s.process == nil {
		return errors.New("StaticSupervisor process is not started")
	}
	return s.process.Send(msg, timeout)
}

func (s *StaticSupervisor) setupProc(link gotp.PID, timeout time.Duration, opts ...gotp.SpawnOpt) <-chan struct{} {
	sig := make(chan struct{})
	s.process = gotp.SpawnLink(link, s.loop(sig), timeout, opts...)
	return sig
}

func (s *StaticSupervisor) startChild(child Supervisable, timeout time.Duration) (reason error) {
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
	child, err := child.StartLink(s.process.ID(), timeout, spec.SpawnOpts...)
	if err != nil {
		return err
	}
	s.registerChild(child)
	return nil
}

func (s *StaticSupervisor) registerChild(supervisable Supervisable) {
	pid := supervisable.ID()
	s.children[pid] = child{supervisable: supervisable}
	if name := supervisable.ChildSpec().Name; name != "" {
		s.childNames[name] = pid
	}
}

func (s *StaticSupervisor) deregisterChild(pid gotp.PID) {
	if child, ok := s.children[pid]; ok {
		spec := child.supervisable.ChildSpec()
		if spec.Name != "" {
			delete(s.childNames, spec.Name)
		}
	}
}

func (s *StaticSupervisor) shouldRestart(pid gotp.PID) (restart bool) {
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

func (s *StaticSupervisor) shouldExit(pid gotp.PID) (exit bool) {
	return s.process.ID() == pid
}

func (s *StaticSupervisor) loop(sig chan struct{}) gotp.RunFn {
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
				slog.Debug("StaticSupervisor.loop: sending exits to children")
				timeout := time.After(s.flags.Shutdown)
			shutdown:
				for pid := range s.children {
					gotp.Send(pid, gotp.NewExit(pid, reason), 0)

					select {
					case msg := <-in:
						if msg, ok := msg.(gotp.Exit); ok {
							if msg.PID() == s.process.ID() {
								continue
							}
							slog.Debug("StaticSupervisor.loop: received exit", "msg", msg)
							s.deregisterChild(msg.PID())
							continue
						}
						slog.Debug("StaticSupervisor.loop: received message", "msg", msg)

					case <-timeout:
						slog.Debug("StaticSupervisor.loop: timeout waiting for children to exit")
						break shutdown
					}
				}
			}
			slog.Debug("StaticSupervisor.loop: exiting")
			if err := s.process.Exit(reason, 0); err != nil {
				slog.Error("StaticSupervisor.loop: failed to exit process", "error", err)
			}
			slog.Debug("StaticSupervisor.loop: exited")
		}()

		slog.Debug("StaticSupervisor.loop: starting children")

		for _, child := range s.specs {
			if err := s.startChild(child, 0); err != nil {
				return err
			}
		}

		slog.Debug("StaticSupervisor.loop: children started")

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
					slog.Debug("StaticSupervisor.loop: input channel closed")
					return nil
				}

				slog.Debug("StaticSupervisor.loop: received message", "msg", msg)
				switch msg := msg.(type) {

				case gotp.Exit:
					s.deregisterChild(msg.PID())
					if s.shouldRestart(msg.PID()) {
						if err := s.startChild(s.children[msg.PID()].supervisable, 0); err != nil {
							slog.Error("DynamicSupervisor.loop: failed to restart child", "error", err)
							hasFailed = true
						}
					} else if s.shouldExit(msg.PID()) {
						slog.Debug("StaticSupervisor.loop: being told to exit")
						return msg.Unwrap()
					}
				}

			case <-ticker.C:
				slog.Debug("StaticSupervisor.loop: checking children")
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
