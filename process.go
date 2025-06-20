package gotp

import (
	"time"

	"github.com/Morgahl/gotp/debug"
)

type Startable interface {
	Start(opts ...SpawnOpt) (Started, error)
}

type Started interface {
	PID() PID
	Send(Msg, time.Duration) error
	SendAfter(Msg, time.Duration) (*time.Timer, error)
}

type SpawnOpt func(*Process)

type RunFn func(*Process) error

type Process struct {
	pid    PID
	linked PID

	mailbox     Mailbox[Msg]
	deregHandle func()
}

func Spawn(fn RunFn, opts ...SpawnOpt) *Process {
	p := build(PID{}, opts)
	p.deregHandle = register(p)
	go p.run(fn)
	return p
}

func SpawnLink(fn RunFn, link PID, opts ...SpawnOpt) *Process {
	p := build(link, opts)
	p.deregHandle = register(p)
	go p.run(fn)
	return p
}

func build(link PID, opts []SpawnOpt) *Process {
	p := &Process{pid: nextPID(), linked: link}
	for _, opt := range opts {
		opt(p)
	}
	if p.mailbox.ch == nil {
		p.mailbox = NewMailbox[Msg](MAILBOX_SIZE)
	}
	return p
}

func (p *Process) run(fn RunFn) {
	var reason error
	defer func() {
		if r := recover(); r != nil {
			reason = debug.Catch(reason, r)
		}
		if p.deregHandle != nil {
			p.deregHandle()
			p.deregHandle = nil
		}
		if !p.linked.IsZero() {
			_ = Send(p.linked, NewExit(p.pid, reason), 0)
		}
	}()
	reason = fn(p)
}

func (p *Process) Exit(reason error, timeout time.Duration) error {
	return p.mailbox.Send(NewExit(p.pid, reason), timeout)
}

func (p *Process) PID() PID {
	return p.pid
}

func (p *Process) Send(msg Msg, timeout time.Duration) error {
	return p.mailbox.Send(msg, timeout)
}

func (p *Process) SendAfter(msg Msg, after time.Duration) (*time.Timer, error) {
	if p.mailbox.ch == nil {
		return nil, NewNotStarted()
	}
	return time.AfterFunc(after, func() {
		_ = p.Send(msg, 0)
	}), nil
}

func (p *Process) Receive() <-chan Msg {
	return p.mailbox.Receive()
}
