package gotp

import (
	"context"
	"fmt"
	"time"
)

type SpawnOpt func(*Process)

type RunFn func(ctx context.Context, proc *Process) error

type Process struct {
	pid    PID
	linked PID

	mailbox     Mailbox[Msg]
	reason      error
	deregHandle func()
}

func Spawn(ctx context.Context, fn RunFn, opts ...SpawnOpt) *Process {
	p := build(PID{}, opts)
	p.deregHandle = register(p)
	ctx = p.setupContext(ctx)
	go p.run(ctx, fn)
	return p
}

func SpawnLink(ctx context.Context, link PID, fn RunFn, opts ...SpawnOpt) *Process {
	p := build(link, opts)
	p.deregHandle = register(p)
	ctx = p.setupContext(ctx)
	go p.run(ctx, fn)
	return p
}

func build(link PID, opts []SpawnOpt) *Process {
	ctx := context.Background()
	p := &Process{pid: nextPID(), linked: link}
	for _, opt := range opts {
		opt(p)
	}
	if p.mailbox.ch == nil {
		p.mailbox = NewMailbox[Msg](ctx, MAILBOX_SIZE)
	}
	return p
}

func (p *Process) setupContext(ctx context.Context) context.Context {
	if !p.linked.IsZero() {
		ctx = context.WithValue(ctx, linkedKey{}, p.linked)
	}
	ctx = context.WithValue(ctx, pidKey{}, p.pid)
	return ctx
}

func (p *Process) run(ctx context.Context, fn RunFn) {
	var reason error
	defer p.cleanup(ctx, &reason)
	reason = fn(ctx, p)
}

func (p *Process) cleanup(ctx context.Context, reason *error) {
	var haveLock bool
	if r := recover(); r != nil {
		p.mailbox.mu.Lock()
		haveLock = true
		if reason != nil {
			p.reason = fmt.Errorf("reason: %v, panic: %v", *reason, r)
		} else {
			p.reason = fmt.Errorf("panic: %v", r)
		}
	}
	if !haveLock {
		p.mailbox.mu.Lock()
	}
	defer p.mailbox.mu.Unlock()
	defer p.deregHandle()
	p.mailbox.Cancel(p.reason)
	if !p.linked.IsZero() {
		Send(ctx, p.linked, NewExit(p.pid, p.reason))
	}
}

func (p *Process) Exit(ctx context.Context, reason error) error {
	p.cleanup(ctx, &reason)
	select {
	case <-p.mailbox.ch:
		return p.reason
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *Process) ID() PID {
	return p.pid
}

func (p *Process) Send(ctx context.Context, msg Msg) error {
	return p.mailbox.Send(ctx, msg)
}

func (p *Process) SendAfter(ctx context.Context, msg Msg, after time.Duration) *time.Timer {
	return time.AfterFunc(after, func() {
		_ = p.Send(ctx, msg)
	})
}

func (p *Process) Receive(match func(Msg) bool, opts ...ReceiveOpt[Msg]) (Msg, bool) {
	return p.mailbox.Receive(match, opts...)
}

type pidKey struct{}

type linkedKey struct{}

func CtxPID(ctx context.Context) (PID, bool) {
	pid, ok := ctx.Value(pidKey{}).(PID)
	return pid, ok
}

func CtxLinked(ctx context.Context) (PID, bool) {
	pid, ok := ctx.Value(linkedKey{}).(PID)
	return pid, ok
}
