package gotp

import (
	"sync"
	"time"
)

const (
	MAILBOX_SIZE     = 100
	DEFAULT_TIMEOUT  = 5 * time.Second
	DEFAULT_SHUTDOWN = 30 * time.Second
)

type MailboxWeak = Mailbox[Msg]

type Atom string

type ReceiveOpt[M any] func(*receiveOpts[M])

type receiveOpts[M any] struct {
	timeout   time.Duration
	onTimeout func() M
}

func WithTimeout[M any](dur time.Duration, cb func() M) ReceiveOpt[M] {
	if dur > 0 && cb == nil {
		cb = func() M {
			var zero M
			return zero
		}
	} else if dur <= 0 && cb != nil {
		panic("WithTimeout: callback function cannot be set without a timeout")
	}

	return func(r *receiveOpts[M]) {
		r.timeout = dur
		r.onTimeout = cb
	}
}

type Mailbox[M Msg] struct {
	mu sync.RWMutex
	ch chan M
}

func NewMailbox[M Msg](bufferSize int) Mailbox[M] {
	return Mailbox[M]{ch: make(chan M, bufferSize)}
}

func (m *Mailbox[M]) Send(msg M, timeout time.Duration) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if timeout > 0 {
		select {
		case <-time.After(timeout):
			return NewTimeout(timeout)
		case m.ch <- msg:
		}
	} else {
		m.ch <- msg
	}
	return nil
}

func (m *Mailbox[M]) Receive() <-chan M {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ch
}
