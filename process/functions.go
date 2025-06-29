package process

import "time"

func Send[M Message](p *Process, m M) {
	p.mailboxLock.Lock()
	p.handleSignal(messageSignal[Message](NO_FLAGS, m))
	p.mailboxLock.Unlock()
}

func Receive[M Message](p *Process, timeout time.Duration) (m M, ok bool) {
	return receive[M](p, timeout)
}

func receive[M Message](p *Process, timeout time.Duration) (M, bool) {
	defer p.maybeGarbageCollect()
	var after <-chan time.Time
	if timeout > 0 {
		after = time.After(timeout)
	}
	p.mailboxLock.Lock()
	for {
		for i, m := range p.mailbox {
			if mt, ok := m.(M); ok {
				p.mailbox[i] = nil
				p.mailboxLock.Unlock()
				return mt, true
			}
		}
		p.mailboxLock.Unlock()
		p.stateLock.RLock()
		switch p.state {
		case STARTING_STATE, STARTED_STATE:
			p.stateLock.RUnlock()
			select {
			case <-after:
				return *new(M), false
			case <-time.After(100 * time.Microsecond):
				continue
			}
		case EXITING_STATE, EXITED_STATE:
			p.stateLock.RUnlock()
			var zero M
			return zero, false
		}
	}
}
