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

	p.mailboxMu.Lock()
	defer p.mailboxMu.Unlock()
	var readOffset int
	for {
		switch p.state {
		case STARTING_STATE, STARTED_STATE:
			for _, m := range p.mailbox[readOffset:] {
				if m == nil {
					readOffset++
					continue
				}
				if mt, ok := m.(M); ok {
					p.mailbox[readOffset] = nil
					p.messageSkips = append(p.messageSkips, readOffset)
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

			for _, m := range p.mailbox[readOffset:] {
				if m == nil {
					readOffset++
					continue
				}
				if mt, ok := m.(M); ok {
					p.mailbox[readOffset] = nil
					p.messageSkips = append(p.messageSkips, readOffset)
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

func Exit[S Sendable](s S, reason error) {
	defer func() { recover() }()
	switch v := any(s).(type) {
	case *Process:
		v.send(exitSignal(no_FLAGS, v.pid, nil, reason))

	case *Ref:
		v.send(exitSignal(no_FLAGS, v.pid, v, reason))

	case PID:
		sendPID(v, exitSignal(no_FLAGS, v, nil, reason))

	case gotp.Atom:
		if pid, found := namedPID(v); found {
			sendPID(pid, exitSignal(no_FLAGS, pid, nil, reason))
		}
	}
}
