package gotp

import (
	"time"
)

const (
	MAILBOX_SIZE     = 16
	DEFAULT_TIMEOUT  = 5 * time.Second
	DEFAULT_SHUTDOWN = 30 * time.Second
)

type MailboxWeak = Mailbox[Msg]

type Atom string
type Mailbox[M Msg] struct {
	ch chan M
}

func NewMailbox[M Msg](bufferSize int) Mailbox[M] {
	return Mailbox[M]{ch: make(chan M, bufferSize)}
}

func (m *Mailbox[M]) Send(msg M) {
	select {
	case m.ch <- msg:
	default:
		// TODO: use an actual mailbox...
		go func() { m.ch <- msg }()
	}
}

func (m *Mailbox[M]) Receive() <-chan M {
	return m.ch
}
