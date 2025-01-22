package gotp

import (
	"context"
	"fmt"
	"sync"
	"time"
)

const (
	MAILBOX_SIZE = 100
	UNLINKED     = 0
)

type PID uint64

func (p PID) String() string {
	return fmt.Sprintf("PID<%d>", p)
}

type SpawnOpt func(*Process)

func Size(size int) SpawnOpt {
	return func(p *Process) {
		if p.ch == nil {
			p.ch = make(chan Msg, size)
		}
	}
}

type RunFn func(ctx context.Context, proc *Process, in <-chan Msg) error

type Process struct {
	pid    PID
	linked PID

	// mu protects everything below treating lifecycle events as full locks and send events as read
	// locks; reason and error
	mu sync.RWMutex
	ch chan Msg

	reason      error
	exit        context.CancelFunc
	deregHandle func()
}

func Spawn(ctx context.Context, fn RunFn, opts ...SpawnOpt) *Process {
	p := build(UNLINKED, opts)
	ctx = p.setupContext(ctx)
	p.deregHandle = register(p)
	go p.run(ctx, fn)
	return p
}

func SpawnLink(ctx context.Context, link PID, fn RunFn, opts ...SpawnOpt) *Process {
	p := build(link, opts)
	ctx = p.setupContext(ctx)
	p.deregHandle = register(p)
	go p.run(ctx, fn)
	return p
}

func build(link PID, opts []SpawnOpt) *Process {
	p := &Process{pid: nextPID(), linked: link}

	for _, opt := range opts {
		opt(p)
	}

	// handle defaults lazily to limit allocs
	if p.ch == nil {
		p.ch = make(chan Msg, MAILBOX_SIZE)
	}

	return p
}

func (p *Process) cleanup(ctx context.Context, reason *error) {
	var haveLock bool
	if r := recover(); r != nil {
		p.mu.Lock()
		haveLock = true
		if reason != nil {
			p.reason = fmt.Errorf("reason: %v, panic: %v", *reason, r)
		} else {
			p.reason = fmt.Errorf("panic: %v", r)
		}
	}
	if !haveLock {
		p.mu.Lock()
	}
	defer p.mu.Unlock()
	defer p.deregHandle()
	p.exit()
	if p.linked != UNLINKED {
		Send(ctx, p.linked, NewExit(p.pid, p.reason))
	}
}

func (p *Process) setupContext(ctx context.Context) context.Context {
	if p.linked != UNLINKED {
		ctx = context.WithValue(ctx, linkedKey{}, p.linked)
	}
	ctx = context.WithValue(ctx, pidKey{}, p.pid)
	ctx, p.exit = context.WithCancel(ctx)
	return ctx
}

func (p *Process) run(ctx context.Context, fn RunFn) {
	var reason error
	defer p.cleanup(ctx, &reason)
	reason = fn(ctx, p, p.ch)
}

func (p *Process) Exit(ctx context.Context, reason error) error {
	p.cleanup(ctx, &reason)
	select {
	case <-p.ch:
		return p.reason
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *Process) ID() PID {
	return p.pid
}

func (p *Process) Send(ctx context.Context, msg Msg) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	select {
	case <-ctx.Done():
		return NewTimeout(ctx.Err())
	case p.ch <- msg:
		return nil
	}
}

func (p *Process) SendAfter(ctx context.Context, msg Msg, after time.Duration) *time.Timer {
	return time.AfterFunc(after, func() {
		p.Send(ctx, msg)
	})
}

type pidKey struct{}

func CtxPID(ctx context.Context) (PID, bool) {
	pid, ok := ctx.Value(pidKey{}).(PID)
	return pid, ok
}

type linkedKey struct{}

func CtxLinked(ctx context.Context) (PID, bool) {
	pid, ok := ctx.Value(linkedKey{}).(PID)
	return pid, ok
}
