package process

import (
	"context"
	"time"

	"github.com/Morgahl/gotp"
)

type Context struct {
	process *process
	ref     Ref
	pidMap  map[PID]Ref
	nameMap map[gotp.Atom]Ref
}

func newContext(process *process) Context {
	return Context{
		process: process,
		ref:     newRef(process),
		pidMap:  make(map[PID]Ref),
		nameMap: make(map[gotp.Atom]Ref),
	}
}

func (c *Context) Ref() Ref {
	return c.ref
}

func (c *Context) PID() PID {
	return c.process.pid
}

func (c *Context) Name() gotp.Atom {
	return c.process.name
}

func (c *Context) Context() context.Context {
	return c.process.context
}

func (c *Context) Sensitive(sensitive bool) {
	if sensitive {
		c.process.flags |= SENSITIVE_FLAG
	} else {
		c.process.flags &^= SENSITIVE_FLAG
	}
}

func (c *Context) TrapExit(trap bool) {
	if trap {
		c.process.flags |= TRAP_EXIT_FLAG
	} else {
		c.process.flags &^= TRAP_EXIT_FLAG
	}
}

func (c *Context) UpdateFlags(fn func(ProcessFlags) ProcessFlags) {
	c.process.flags = fn(c.process.flags)
}

func (c *Context) Send(msg Message) {
	c.process.send(messageSignal(no_FLAGS, msg))
}

func (c *Context) SendAfter(msg Message, delay time.Duration) *time.Timer {
	return time.AfterFunc(delay, func() { c.process.send(messageSignal(no_FLAGS, msg)) })
}

// ProcessPending processes any pending signals for the current process. Since we are ultimately still limited by a
// channel while operating within a receive loop we need a way to process the current pending signals in cases where
// deadlock might be found.
func (c *Context) ProcessPending() {
	switch c.process.state {
	case STARTING_STATE, STARTED_STATE:
		select {
		case s, ok := <-c.process.signalChan:
			if !ok {
				return
			}
			c.process.handleSignal(s)
		default:
			return
		}
	}
}

func (c *Context) Exit(reason error) {
	c.process.send(exitSignal(no_FLAGS, c.process.pid, newRef(c.process), reason))
}

func ContextSend[S Sendable](pctx Context, to S, m Message) {
	defer func() { recover() }()
	switch v := any(to).(type) {
	case Ref:
		v.send(messageSignal(no_FLAGS, m))

	case PID:
		if ref, found := pctx.pidMap[v]; found {
			if ref.IsValid() {
				ref.send(messageSignal(no_FLAGS, m))
				return
			}
			delete(pctx.pidMap, v)
		}
		if ref, found := pidRef(v); found && ref.IsValid() {
			pctx.pidMap[v] = ref
			ref.send(messageSignal(no_FLAGS, m))
		}

	case gotp.Atom:
		if ref, found := pctx.nameMap[v]; found {
			if ref.IsValid() {
				ref.send(messageSignal(no_FLAGS, m))
				return
			}
			delete(pctx.nameMap, v)
		}
		if ref, found := namedRef(v); found && ref.IsValid() {
			pctx.nameMap[v] = ref
			ref.send(messageSignal(no_FLAGS, m))
		}
	}
}

func ContextSendAfter[S Sendable](pctx Context, s S, m Message, delay time.Duration) *time.Timer {
	return time.AfterFunc(delay, func() { ContextSend(pctx, s, m) })
}
