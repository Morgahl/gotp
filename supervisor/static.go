package supervisor

import (
	"context"
	"errors"
	"fmt"
	"log"
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

func (s *StaticSupervisor) StartLink(ctx context.Context, link gotp.PID, opts ...gotp.SpawnOpt) (Supervisable, error) {
	log.Printf("StaticSupervisor.StartLink: %v", link)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-s.setupProc(ctx, link, opts...):
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

func (s *StaticSupervisor) StartChild(child Supervisable) error {
	return errors.New("StaticSupervisor does not support dynamic child addition")
}

func (s *StaticSupervisor) StopChild(pid gotp.PID) error {
	return errors.New("StaticSupervisor does not support dynamic child removal")
}

func (s *StaticSupervisor) setupProc(ctx context.Context, link gotp.PID, opts ...gotp.SpawnOpt) <-chan struct{} {
	sig := make(chan struct{})
	s.process = gotp.SpawnLink(ctx, link, s.loop(sig), opts...)
	return sig
}

func (s *StaticSupervisor) startChild(ctx context.Context, child Supervisable) (reason error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("StaticSupervisor.startChild: panic: %v", r)
			if reason == nil {
				reason = fmt.Errorf("panic: %v", r)
			} else {
				reason = fmt.Errorf("reason: %w, panic: %v", reason, r)
			}
		}
	}()
	spec := child.ChildSpec()
	ctx, cancel := context.WithCancel(ctx)
	child, err := child.StartLink(ctx, s.process.ID(), spec.SpawnOpts...)
	if err != nil {
		cancel()
		return err
	}
	s.registerChild(child, cancel)
	return nil
}

func (s *StaticSupervisor) registerChild(supervisable Supervisable, cancel func()) {
	pid := supervisable.ID()
	cs := child{
		supervisable: supervisable,
		cancel:       cancel,
	}
	s.children[pid] = cs

	if name := supervisable.ChildSpec().Name; name != "" {
		s.childNames[name] = pid
	}
}

func (s *StaticSupervisor) deregisterChild(pid gotp.PID) {
	if child, ok := s.children[pid]; ok {
		spec := child.supervisable.ChildSpec()
		if child.cancel != nil {
			child.cancel()
			child.cancel = nil
		}
		if spec.Name != "" {
			delete(s.childNames, spec.Name)
		}
	}
}

func (s *StaticSupervisor) shouldRestart(pid gotp.PID) (restart bool) {
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

// We reimplement the loop function using Mailbox.Receive(match, ...) Mailbox.Chan no longer exists and should not be used.
func (s *StaticSupervisor) loop(sig chan struct{}) gotp.RunFn {
	return func(ctx context.Context, p *gotp.Process) (reason error) {
		defer func() {
			if r := recover(); r != nil {
				if reason == nil {
					reason = fmt.Errorf("panic: %v", r)
				} else {
					reason = fmt.Errorf("reason: %w, panic: %v", reason, r)
				}
			}
			for _, cs := range s.children {
				if cs.cancel != nil {
					cs.cancel()
					cs.cancel = nil
				}
			}
			var msg gotp.Msg
			var ok bool
			for msg, ok = p.Receive(matchExit); ok; msg, ok = p.Receive(matchExit) {
				if msg, ok := msg.(gotp.Exit); ok {
					log.Printf("StaticSupervisor.loop: received exit: %v", msg)
					s.deregisterChild(msg.PID())
				}
			}
			s.process.Exit(ctx, reason)
		}()

		log.Printf("StaticSupervisor.loop: starting children")

		for _, child := range s.specs {
			if err := s.startChild(ctx, child); err != nil {
				return err
			}
		}

		log.Printf("StaticSupervisor.loop: children started")

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
			case <-ctx.Done():
				log.Printf("StaticSupervisor.loop: context done")
				return ctx.Err()

			case <-ticker.C:
				log.Printf("StaticSupervisor.loop: checking children")
				if !hasFailed {
					continue
				}

				var newFailed bool
				for pid := range s.children {
					if _, exists := s.children[pid]; !exists {
						if s.shouldRestart(pid) {
							if err := s.startChild(ctx, s.children[pid].supervisable); err != nil {
								log.Printf("StaticSupervisor.loop: failed to restart child: %v", err)
								newFailed = true
							}
						}
					}
				}
				hasFailed = newFailed

			default:
				if msg, ok := p.Receive(matchExit); ok {
					log.Printf("StaticSupervisor.loop: received message: %T(%v)", msg, msg)
					switch msg := msg.(type) {
					case gotp.Exit:
						s.deregisterChild(msg.PID())
						if s.shouldRestart(msg.PID()) {
							if err := s.startChild(ctx, s.children[msg.PID()].supervisable); err != nil {
								log.Printf("StaticSupervisor.loop: failed to restart child: %v", err)
								hasFailed = true
							}
						}
					}
				}
			}
		}
	}
}
