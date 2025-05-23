package gotp

import (
	"context"
	"log"
	"slices"
	"sync"
	"time"
)

const (
	MAILBOX_SIZE = 100
	UNLINKED     = 0
)

type MailboxWeak = Mailbox[Msg]

type Atom string

type ReceiveOpt[M any] func(*receiveOpts[M])

type receiveOpts[M any] struct {
	timeout   time.Duration
	onTimeout func() M
}

func WithTimeout[M any](dur time.Duration, cb func() M) ReceiveOpt[M] {
	return func(r *receiveOpts[M]) {
		r.timeout = dur
		r.onTimeout = cb
	}
}

type Mailbox[M Msg] struct {
	context.Context
	cancel    context.CancelCauseFunc
	ch        chan M
	unhandled []M
	mu        sync.RWMutex
}

func NewMailbox[M Msg](parent context.Context, bufferSize int) Mailbox[M] {
	ctx, cancel := context.WithCancelCause(parent)
	return Mailbox[M]{Context: ctx, cancel: cancel, ch: make(chan M, bufferSize)}
}

func (m *Mailbox[M]) Cancel(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cancel(err)
}

func (m *Mailbox[M]) Chan() <-chan M {
	log.Println("Mailbox.Chan used (DEPRECATED IN FAVOR OF RECEIVE)")
	return m.ch
}

func (m *Mailbox[M]) Send(ctx context.Context, msg M) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	select {
	case <-ctx.Done():
		return NewTimeout(ctx.Err())
	case m.ch <- msg:
		return nil
	}
}

// Receive attempts to match a message using the provided match function.
// If no match is found immediately, it continues listening until a match is received,
// the context is cancelled (Inbox.Done), or the timeout is reached.
// This is designed to support cascading if-else matching logic:
//
//	if val, ok := inbox.Receive(ctx, match1); ok { ... }
//	else if val, ok := inbox.Receive(ctx, match2); ok { ... }
//	else if ctx.Err() != nil { ... } // safe fallback when all fails.
func (m *Mailbox[M]) Receive(match func(M) bool, opts ...ReceiveOpt[M]) (M, bool) {
	var zero M
	var config receiveOpts[M]
	for _, opt := range opts {
		opt(&config)
	}
	for idx := 0; idx < len(m.unhandled); {
		select {
		case <-m.Done():
			return zero, false
		default:

			if um := m.unhandled[idx]; match(um) {
				m.unhandled = slices.Delete(m.unhandled, idx, idx+1)
				return um, true
			}
			idx++
		}
	}
	timer := time.NewTimer(config.timeout)
	defer timer.Stop()
	for {
		select {
		case <-m.Done():
			var zero M
			return zero, false
		case <-timer.C:
			if config.onTimeout != nil {
				return config.onTimeout(), true
			}
			var zero M
			return zero, false
		case msg := <-m.ch:
			if match(msg) {
				return msg, true
			}
			m.unhandled = append(m.unhandled, msg)
		}
	}
}
