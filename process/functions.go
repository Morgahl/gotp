package process

import (
	"time"

	"github.com/Morgahl/gotp"
)

type Sendable interface {
	*Process | *Ref | PID | gotp.Atom
}

func Send[S Sendable](s S, m Message) {
	defer func() { recover() }()
	switch v := any(s).(type) {
	case *Process:
		v.send(messageSignal(no_FLAGS, m))

	case *Ref:
		v.send(messageSignal(no_FLAGS, m))

	case PID:
		sendPID(v, m)

	case gotp.Atom:
		sendNamed(v, m)
	}
}

func SendAfter[S Sendable](s S, m Message, delay time.Duration) *time.Timer {
	return time.AfterFunc(delay, func() { Send(s, m) })
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
