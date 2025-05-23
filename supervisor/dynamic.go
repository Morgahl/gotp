package supervisor

import (
	"context"
	"errors"
	"fmt"
	"log"
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
	ctx      context.Context
	children map[gotp.PID]child
}

func Dynamic(flags Flags) *DynamicSupervisor {
	return &DynamicSupervisor{
		flags:    flags.ApplyDefaults(),
		children: make(map[gotp.PID]child),
	}
}

func (s *DynamicSupervisor) StartLink(ctx context.Context, link gotp.PID, opts ...gotp.SpawnOpt) (Supervisable, error) {
	log.Printf("DynamicSupervisor.StartLink: %v", link)
	s.ctx = ctx
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-s.setupProc(link, opts...):
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

func (s *DynamicSupervisor) StartChild(child Supervisable) error {
	// TODO: send a message to the Supervisor process to start the child so that there is no need to
	// TODO: use a mutex to protect the maps
	if s.process == nil {
		return errors.New("DynamicSupervisor not started")
	}
	return s.startChild(s.ctx, child)
}

func (s *DynamicSupervisor) StopChild(pid gotp.PID) error {
	// TODO: send a message to the Supervisor process to stop the child so that there is no need to
	// TODO: use a mutex to protect the maps
	if s.process == nil {
		return errors.New("DynamicSupervisor not started")
	}
	s.stopChild(pid)
	return nil
}

func (s *DynamicSupervisor) setupProc(link gotp.PID, opts ...gotp.SpawnOpt) <-chan struct{} {
	sig := make(chan struct{})
	s.process = gotp.SpawnLink(s.ctx, link, s.loop(sig), opts...)
	return sig
}

func (s *DynamicSupervisor) startChild(ctx context.Context, child Supervisable) (reason error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("DynamicSupervisor.startChild: panic: %v", r)
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

func (s *DynamicSupervisor) stopChild(pid gotp.PID) {
	if child, exists := s.children[pid]; exists {
		child.cancel()
		delete(s.children, pid)
	}
}

func (s *DynamicSupervisor) registerChild(supervisable Supervisable, cancel func()) {
	s.children[supervisable.ID()] = child{
		supervisable: supervisable,
		cancel:       cancel,
		restart:      restart{},
	}
}

func (s *DynamicSupervisor) deregisterChild(pid gotp.PID) {
	if child, ok := s.children[pid]; ok {
		child.cancel()
		delete(s.children, pid)
	}
}

func (s *DynamicSupervisor) shouldRestart(pid gotp.PID) (restart bool) {
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

func (s *DynamicSupervisor) loop(sig chan struct{}) func(context.Context, *gotp.Process) error {
	return func(ctx context.Context, p *gotp.Process) (reason error) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("DynamicSupervisor.loop: panic: %v", r)
				if reason == nil {
					reason = fmt.Errorf("panic: %v", r)
				} else {
					reason = fmt.Errorf("reason: %w, panic: %v", reason, r)
				}
			}
			for _, cs := range s.children {
				cs.cancel()
			}
			s.process.Exit(ctx, reason)
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
			case <-ctx.Done():
				log.Printf("DynamicSupervisor.loop: context done")
				return ctx.Err()

			case <-ticker.C:
				log.Printf("DynamicSupervisor.loop: checking children")
				if !hasFailed {
					continue
				}

				var newFailed bool
				for pid := range s.children {
					if _, exists := s.children[pid]; !exists {
						if s.shouldRestart(pid) {
							if err := s.startChild(ctx, s.children[pid].supervisable); err != nil {
								log.Printf("DynamicSupervisor.loop: failed to restart child: %v", err)
								newFailed = true
							}
						}
					}
				}
				hasFailed = newFailed

			default:
				if msg, ok := p.Receive(matchExit); ok {
					msg := msg.(gotp.Exit)
					log.Printf("DynamicSupervisor.loop: received exit: %v", msg)
					s.deregisterChild(msg.PID())
					if s.shouldRestart(msg.PID()) {
						if err := s.startChild(ctx, s.children[msg.PID()].supervisable); err != nil {
							log.Printf("DynamicSupervisor.loop: failed to restart child: %v", err)
							hasFailed = true
						}
					}
				} else if msg, ok := p.Receive(matchStartChild); ok {
					msg := msg.(startChild)
					log.Printf("DynamicSupervisor.loop: received start child: %v", msg)
					if err := s.startChild(ctx, msg.child); err != nil {
						log.Printf("DynamicSupervisor.loop: failed to start child: %v", err)
						hasFailed = true
					}
				} else if msg, ok := p.Receive(matchStopChild); ok {
					msg := msg.(stopChild)
					log.Printf("DynamicSupervisor.loop: received stop child: %v", msg)
					s.stopChild(msg.pid)
				} else if ctx.Err() != nil {
					log.Printf("DynamicSupervisor.loop: context done")
					return ctx.Err()
				} else {
					log.Printf("DynamicSupervisor.loop: no message")
					// no message received, continue the loop
					continue
				}
			}
		}
	}
}

func matchStartChild(msg gotp.Msg) bool {
	if _, ok := msg.(startChild); ok {
		return true
	}
	return false
}

func matchStopChild(msg gotp.Msg) bool {
	if _, ok := msg.(stopChild); ok {
		return true
	}
	return false
}
