package gotp

import (
	"fmt"
	"log/slog"
	"time"
)

type Running interface {
	PID() PID
	Send(Msg, time.Duration) error
	SendAfter(Msg, time.Duration) (*time.Timer, error)
	Receive() <-chan Msg
	Exit(error, time.Duration) error
	Exited() bool
}

type SpawnOpt func(*Process)

type RunFn func(*Process, <-chan Msg) error

type Process struct {
	pid    PID
	linked PID

	mailbox     Mailbox[Msg]
	reason      error
	deregHandle func()
}

func Spawn(fn RunFn, timeout time.Duration, opts ...SpawnOpt) *Process {
	p := build(PID{}, opts)
	p.deregHandle = register(p)
	go p.run(fn)
	return p
}

func SpawnLink(link PID, fn RunFn, timeout time.Duration, opts ...SpawnOpt) *Process {
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
	defer p.cleanup(&reason, 0)
	reason = fn(p, p.mailbox.ch)
}

func (p *Process) cleanup(reason *error, timeout time.Duration) error {
	p.mailbox.mu.Lock()
	defer p.mailbox.mu.Unlock()
	if p.deregHandle == nil {
		slog.Debug("Process already deregistered", "pid", p.pid)
		return nil
	}

	defer func() {
		p.deregHandle()
		p.deregHandle = nil
	}()
	if r := recover(); r != nil {
		slog.Error("Process panicked", "pid", p.pid, "reason", r)
		if reason != nil {
			p.reason = fmt.Errorf("reason: %v, panic: %v", *reason, r)
		} else {
			p.reason = fmt.Errorf("panic: %v", r)
		}
	} else if *reason != nil {
		p.reason = *reason
	}

	if !p.linked.IsZero() {
		slog.Debug("Process sending exit to linked process", "pid", p.pid, "linked", p.linked)
		return Send(p.linked, NewExit(p.pid, p.reason), timeout)
	}
	slog.Debug("Process exiting", "pid", p.pid, "reason", p.reason)
	return nil
}

func (p *Process) Exit(reason error, timeout time.Duration) error {
	slog.Debug("Process exiting: calling cleanup", "pid", p.pid)
	return p.cleanup(&reason, timeout)
}

func (p *Process) Exited() bool {
	return p.reason != nil
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

func (p *Process) Reason() error {
	if p.reason != nil {
		return p.reason
	}
	return fmt.Errorf("process %s not exited", p.pid)
}
