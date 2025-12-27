package process

import (
	"context"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/term"
)

type Context struct {
	process *process
	ref     Ref
}

func newContext(process *process) Context {
	return Context{
		process: process,
		ref:     newRef(process),
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

func (c *Context) Send(msg term.Term) {
	c.process.send(messageSignal(no_FLAGS, msg))
}

func (c *Context) SendAfter(msg term.Term, delay time.Duration) *time.Timer {
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
