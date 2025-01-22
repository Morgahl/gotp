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

	process *gotp.Process
	ctx     context.Context
	// TODO: merge same keyed maps
	childPIDs   map[gotp.PID]Supervisable
	childCancel map[gotp.PID]context.CancelFunc
	restarts    map[gotp.PID]restart
}

func Dynamic(flags Flags) *DynamicSupervisor {
	return &DynamicSupervisor{
		flags:       flags.ApplyDefaults(),
		childPIDs:   map[gotp.PID]Supervisable{},
		childCancel: map[gotp.PID]context.CancelFunc{},
		restarts:    map[gotp.PID]restart{},
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
	if cancel, exists := s.childCancel[pid]; exists {
		cancel()
		delete(s.childCancel, pid)
	}
}

func (s *DynamicSupervisor) registerChild(child Supervisable, cancel func()) {
	pid := child.ID()
	s.childPIDs[pid] = child
	s.childCancel[pid] = cancel
	s.restarts[pid] = restart{}
}

func (s *DynamicSupervisor) deregisterChild(pid gotp.PID) {
	if cancel, ok := s.childCancel[pid]; ok {
		cancel()
		delete(s.childCancel, pid)
	}
	delete(s.childPIDs, pid)
	delete(s.restarts, pid)
}

func (s *DynamicSupervisor) shouldRestart(pid gotp.PID) (restart bool) {
	r, ok := s.restarts[pid]
	if !ok {
		return true
	}

	r.count++

	if r.count == 1 {
		r.at = time.Now()
		restart = true
	} else if r.count <= s.flags.MaxRestarts {
		restart = true
	} else if time.Since(r.at) > s.flags.ResetPeriod {
		r.count = 1
		r.at = time.Now()
		restart = true
	} else {
		restart = false
	}

	if restart {
		s.restarts[pid] = r
	}

	return restart
}

func (s *DynamicSupervisor) loop(sig chan struct{}) func(context.Context, *gotp.Process, <-chan gotp.Msg) error {
	return func(ctx context.Context, _ *gotp.Process, in <-chan gotp.Msg) (reason error) {
		defer func() {
			if r := recover(); r != nil {
				if reason == nil {
					reason = fmt.Errorf("panic: %v", r)
				} else {
					reason = fmt.Errorf("reason: %w, panic: %v", reason, r)
				}
			}
			for _, cancel := range s.childCancel {
				cancel()
			}
			for msg := range in {
				if msg, ok := msg.(gotp.Exit); ok {
					s.deregisterChild(msg.PID())
				}
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

			case msg := <-in:
				log.Printf("DynamicSupervisor.loop: received message: %T(%v)", msg, msg)
				switch msg := msg.(type) {
				case gotp.Exit:
					s.deregisterChild(msg.PID())
					if s.shouldRestart(msg.PID()) {
						if err := s.startChild(ctx, s.childPIDs[msg.PID()]); err != nil {
							log.Printf("DynamicSupervisor.loop: failed to restart child: %v", err)
							hasFailed = true
						}
					}

				case startChild:
					if err := s.startChild(ctx, msg.child); err != nil {
						log.Printf("DynamicSupervisor.loop: failed to start child: %v", err)
						hasFailed = true
					}

				case stopChild:
					s.stopChild(msg.pid)
				}

			case <-ticker.C:
				log.Printf("DynamicSupervisor.loop: checking children")
				if !hasFailed {
					continue
				}

				var newFailed bool
				for pid := range s.childPIDs {
					if _, exists := s.childCancel[pid]; !exists {
						if s.shouldRestart(pid) {
							if err := s.startChild(ctx, s.childPIDs[pid]); err != nil {
								log.Printf("DynamicSupervisor.loop: failed to restart child: %v", err)
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
