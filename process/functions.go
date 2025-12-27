package process

import (
	"context"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/term"
)

type Sendable interface {
	Ref | PID | gotp.Atom
}

func Send[S Sendable](to S, m term.Term) {
	defer func() { recover() }()
	switch v := any(to).(type) {
	case Ref:
		v.send(messageSignal(no_FLAGS, m))

	case PID:
		sendPID(v, m)

	case gotp.Atom:
		sendNamed(v, m)
	}
}

func SendAfter[S Sendable](s S, m term.Term, delay time.Duration) *time.Timer {
	return time.AfterFunc(delay, func() { Send(s, m) })
}

func ReceiveWithTimeout[M term.Term](pctx Context, timeout time.Duration) (M, bool, error) {
	var after <-chan time.Time
	if timeout > 0 {
		after = time.After(timeout)
	}
	return receive[M](pctx, after)
}

func ReceiveContext[M term.Term](pctx Context, ctx context.Context) (M, bool, error) {
	m, ok, err := receive[M](pctx, ctx.Done())
	if err == nil {
		err = context.Cause(ctx)
	}
	return m, ok, err
}

func receive[M term.Term, D any](pctx Context, done <-chan D) (_ M, _ bool, reason error) {
	var readOffset int
	var messageSkipOffset int
	defer pctx.process.maybeGarbageCollect()
	switch pctx.process.state {
	case STARTING_STATE, STARTED_STATE:
		select {
		case s, ok := <-pctx.process.signalChan:
			if !ok {
				goto EXIT
			}
			pctx.process.handleSignal(s)
		case <-done:
			goto EXIT
		default:
			goto PROCESS_MESSAGES
		}

	case EXITING_STATE, EXITED_STATE:
		goto EXIT
	}

PROCESS_MESSAGES:
	switch pctx.process.state {
	case STARTING_STATE, STARTED_STATE:
		for n, m := range pctx.process.mailbox[readOffset:] {
			if messageSkipOffset < len(pctx.process.messageSkips) && pctx.process.messageSkips[messageSkipOffset] == readOffset+n {
				messageSkipOffset++
				readOffset++
				continue
			}
			if mt, ok := m.(M); ok {
				pctx.process.mailbox[readOffset] = nil
				pctx.process.messageSkips = append(pctx.process.messageSkips, readOffset)
				return mt, true, nil
			}
			readOffset++
		}

	case EXITING_STATE, EXITED_STATE:
		goto EXIT
	}

	select {
	case s, ok := <-pctx.process.signalChan:
		if !ok {
			goto EXIT
		}
		pctx.process.handleSignal(s)
		goto PROCESS_MESSAGES
	case <-done:
		goto EXIT
	}

EXIT:
	var zero M
	if pctx.process.state == EXITING_STATE || pctx.process.state == EXITED_STATE {
		return zero, false, pctx.process.exitReason
	}
	return zero, false, nil
}

func Exit[S Sendable](to S, reason error) {
	defer func() { recover() }()
	switch v := any(to).(type) {
	case Ref:
		v.send(exitSignal(no_FLAGS, v.pid, v, reason))

	case PID:
		sendPID(v, exitSignal(no_FLAGS, v, Ref{}, reason))

	case gotp.Atom:
		sendNamed(v, exitSignal(no_FLAGS, PIDZero(), Ref{}, reason))
	}
}

const ErrBadArg gotp.Atom = "badarg"

func Register(name gotp.Atom, ref Ref) error {
	if name == "undefined" {
		// TODO: maybe better error here?
		return ErrBadArg
	} else if !ref.IsValid() {
		// TODO: maybe better error here?
		return ErrBadArg
	} else if err := nameTree.Store(name.String(), ref); err != nil {
		// TODO: maybe better error here?
		return ErrBadArg
	}
	return nil
}

func WhereIs(name gotp.Atom) Ref {
	return namedRef(name)
}
