package gotp

import (
	"fmt"
	"time"

	"github.com/Morgahl/gotp/debug"
)

type Startable interface {
	Start(opts ...SpawnOpt) (Started, error)
}

type Started interface {
	ID() PID
	Send(Msg)
	SendAfter(Msg, time.Duration) *time.Timer
}

type RunFn func(*Process) error

type Process struct {
	pid    PID
	linked PID
	runFn  RunFn

	mailbox     Mailbox[Msg]
	mbSize      int
	deregHandle func()
}

func Spawn(fn RunFn, opts ...SpawnOpt) *Process {
	p := build(opts)
	p.runFn = fn
	return p
}

func (p *Process) Start() {
	if p.deregHandle != nil {
		panic("Process already started")
	}
	p.deregHandle = register(p)
	go p.run()
}

// func Spawn(fn RunFn, opts ...SpawnOpt) *Process {
// 	p := build(opts)
// 	p.runFn = fn
// 	p.deregHandle = register(p)
// 	go p.run()
// 	return p
// }

// func SpawnLink(fn RunFn, link PID, opts ...SpawnOpt) *Process {
// 	p := build(append(opts, Link(link)))
// 	p.runFn = fn
// 	p.deregHandle = register(p)
// 	go p.run()
// 	return p
// }

func build(opts []SpawnOpt) *Process {
	p := &Process{
		pid:    nextPID(),
		mbSize: MAILBOX_SIZE,
	}
	for _, opt := range opts {
		opt(p)
	}
	p.mailbox = NewMailbox[Msg](p.mbSize)
	return p
}

func (p *Process) run() {
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
			Send(p.linked, NewExit(p.pid, reason))
		}
	}()
	reason = p.runFn(p)
}

func (p Process) String() string {
	return fmt.Sprintf("Process[%T](pid: %s, linked: %s, mailbox size: %d, deregHandle: %t)",
		p, p.pid, p.linked, p.mbSize, p.deregHandle != nil)
}

func (p *Process) Exit(reason error) {
	p.mailbox.Send(NewExit(p.pid, reason))
}

func (p *Process) ID() PID {
	return p.pid
}

func (p *Process) Send(msg Msg) {
	p.mailbox.Send(msg)
}

func (p *Process) SendAfter(msg Msg, after time.Duration) *time.Timer {
	return time.AfterFunc(after, func() {
		p.mailbox.Send(msg)
	})
}

func (p *Process) Receive() <-chan Msg {
	return p.mailbox.Receive()
}
