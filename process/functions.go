package process

import (
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/debug"
)

type Sendable interface {
	*Process | *Ref | PID | gotp.Atom
}

func Send[S Sendable](p S, m Message) {
	switch v := any(p).(type) {
	case *Process:
		// TODO: just one of these should be used, at the top level
		defer func() { recover() }()
		v.send(messageSignal(NO_FLAGS, m))

	case *Ref:
		// TODO: just one of these should be used, at the top level
		defer func() { recover() }()
		v.send(messageSignal(NO_FLAGS, m))

	case PID, gotp.Atom:
		debug.Throw("process.Send: not implemented for PID or gotp.Atom")
		// TODO: this likely requires some additional node local handling of PID allocation as well
		// TODO: as name registration
	}
}

func ReceiveWithTimeout[M Message](p *Process, timeout time.Duration) (M, bool) {
	p.stateLock.RLock()
	defer p.stateLock.RUnlock()
	defer p.maybeGarbageCollect()
	var after <-chan time.Time
	if timeout > 0 {
		after = time.After(timeout)
	}
	readOffset := 0

	p.mailboxMu.Lock()
	defer p.mailboxMu.Unlock()
	for {
		switch p.state {
		case STARTING_STATE, STARTED_STATE:
			for i, m := range p.mailbox[readOffset:] {
				if mt, ok := m.(M); ok {
					p.mailbox[i] = nil
					return mt, true
				}
				readOffset++
			}

			select {
			case s := <-p.signalChan:
				p.handleSignal(s)
				continue
			case <-after:
				var zero M
				return zero, false
			}

		case EXITING_STATE, EXITED_STATE:
			var zero M
			return zero, false
		}
	}
}

func Receive[M Message](p *Process) (M, bool) {
	p.stateLock.RLock()
	defer p.stateLock.RUnlock()
	defer p.maybeGarbageCollect()
	p.mailboxMu.Lock()
	defer p.mailboxMu.Unlock()

	var readOffset int
	for {
		switch p.state {
		case STARTING_STATE, STARTED_STATE:
			for i, m := range p.mailbox[readOffset:] {
				if mt, ok := m.(M); ok {
					p.mailbox[i] = nil
					return mt, true
				}
				readOffset++
			}

			select {
			case s := <-p.signalChan:
				p.handleSignal(s)
				continue
			default:
				var zero M
				return zero, false
			}

		case EXITING_STATE, EXITED_STATE:
			var zero M
			return zero, false
		}
	}
}
