package process

import (
	"context"
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
		// TODO: this currently contends a global mutex, we should consider a more efficient way to send messages to PIDs
		sendPID(v, m)

	case gotp.Atom:
		// TODO: this currently contends a global mutex, we should consider a more efficient way to send messages to named processes
		sendNamed(v, m)
	}
}

func SendAfter[S Sendable](s S, m Message, delay time.Duration) *time.Timer {
	return time.AfterFunc(delay, func() { Send(s, m) })
}

func ReceiveWithTimeout[M Message](p *Process, timeout time.Duration) (M, bool, error) {
	var after <-chan time.Time
	if timeout > 0 {
		after = time.After(timeout)
	}
	return receive[M](p, after)
}

func ReceiveContext[M Message](p *Process, ctx context.Context) (M, bool, error) {
	m, ok, err := receive[M](p, ctx.Done())
	if err == nil {
		err = context.Cause(ctx)
	}
	return m, ok, err
}

func receive[M Message, D any](p *Process, done <-chan D) (_ M, _ bool, reason error) {
	var readOffset int
	var messageSkipOffset int
	p.mailboxMu.Lock()
	defer p.mailboxMu.Unlock()
	defer p.maybeGarbageCollect()
	// yieldAfter := len(p.mailbox) + 1
	// for i := 0; i < yieldAfter; i++ {
	// for {
	switch p.state {
	case STARTING_STATE, STARTED_STATE:
		select {
		case s, ok := <-p.signalChan:
			if !ok {
				goto EXIT
			}
			p.handleSignal(s)
		case <-done:
			goto EXIT
		default:
			goto PROCESS_MESSAGES
		}

	case EXITING_STATE, EXITED_STATE:
		goto EXIT
	}
	// }

PROCESS_MESSAGES:
	switch p.state {
	case STARTING_STATE, STARTED_STATE:
		for n, m := range p.mailbox[readOffset:] {
			if messageSkipOffset < len(p.messageSkips) && p.messageSkips[messageSkipOffset] == readOffset+n {
				messageSkipOffset++
				readOffset++
				continue
			}
			if mt, ok := m.(M); ok {
				p.mailbox[readOffset] = nil
				p.messageSkips = append(p.messageSkips, readOffset)
				return mt, true, nil
			}
			readOffset++
		}

	case EXITING_STATE, EXITED_STATE:
		goto EXIT
	}

	select {
	case s, ok := <-p.signalChan:
		if !ok {
			goto EXIT
		}
		p.handleSignal(s)
		goto PROCESS_MESSAGES
	case <-done:
		goto EXIT
	}
EXIT:
	var zero M
	if p.state == EXITING_STATE || p.state == EXITED_STATE {
		return zero, false, p.exitReason
	}
	return zero, false, nil
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
