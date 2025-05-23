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

type restart struct {
	count uint
	at    time.Time
}

type StaticSupervisor struct {
	flags    Flags
	children []Supervisable

	process *gotp.Process
	// TODO: merge same keyed maps
	childPIDs   map[gotp.PID]Supervisable
	childNames  map[string]gotp.PID
	childCancel map[gotp.PID]context.CancelFunc
	restarts    map[gotp.PID]restart
}

func Static(flags Flags, children ...Supervisable) *StaticSupervisor {
	return &StaticSupervisor{
		flags:       flags.ApplyDefaults(),
		children:    children,
		childPIDs:   make(map[gotp.PID]Supervisable, len(children)),
		childNames:  make(map[string]gotp.PID, len(children)),
		childCancel: make(map[gotp.PID]context.CancelFunc, len(children)),
		restarts:    make(map[gotp.PID]restart, len(children)),
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

func (s *StaticSupervisor) registerChild(child Supervisable, cancel func()) {
	pid := child.ID()
	spec := child.ChildSpec()
	s.childPIDs[pid] = child
	if spec.Name != "" {
		s.childNames[spec.Name] = pid
	}
	s.childCancel[pid] = cancel
	s.restarts[pid] = restart{}
}

func (s *StaticSupervisor) deregisterChild(pid gotp.PID) {
	if child, ok := s.childPIDs[pid]; ok {
		spec := child.ChildSpec()
		if cancel, ok := s.childCancel[pid]; ok {
			cancel()
			delete(s.childCancel, pid)
		}
		if spec.Name != "" {
			delete(s.childNames, spec.Name)
		}
		delete(s.childPIDs, pid)
		delete(s.restarts, pid)
	}
}

func (s *StaticSupervisor) shouldRestart(pid gotp.PID) (restart bool) {
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

func (s *StaticSupervisor) loop(sig chan struct{}) func(context.Context, *gotp.Process, *gotp.Mailbox[gotp.Msg]) error {
	return func(ctx context.Context, _ *gotp.Process, in *gotp.Mailbox[gotp.Msg]) (reason error) {
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
			for msg := range in.Chan() {
				if msg, ok := msg.(gotp.Exit); ok {
					log.Printf("StaticSupervisor.loop: received exit: %v", msg)
					s.deregisterChild(msg.PID())
				}
			}
			s.process.Exit(ctx, reason)
		}()

		log.Printf("StaticSupervisor.loop: starting children")

		for _, child := range s.children {
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

			case msg := <-in.Chan():
				log.Printf("StaticSupervisor.loop: received message: %v", msg)
				switch msg := msg.(type) {

				case gotp.Exit:
					s.deregisterChild(msg.PID())
					if s.shouldRestart(msg.PID()) {
						if err := s.startChild(ctx, s.childPIDs[msg.PID()]); err != nil {
							log.Printf("DynamicSupervisor.loop: failed to restart child: %v", err)
							hasFailed = true
						}
					}
				}

			case <-ticker.C:
				log.Printf("StaticSupervisor.loop: checking children")
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
