package process

import (
	"errors"
	"sync"
	"time"

	"github.com/Morgahl/gotp/debug"
)

type Ref struct {
	pid    PID
	sendFn func(s signal[Message])
}

func (r *Ref) IsValid() bool {
	return r.sendFn != nil
}

func (r *Ref) Send(m Message) (err error) {
	return r.send(messageSignal(NO_FLAGS, m))
}

func (r *Ref) send(s signal[Message]) (err error) {
	defer func() {
		err = debug.Recover(recover(), "*Ref.send", err)
	}()
	if r.sendFn == nil {
		return errors.New("*Ref.send: bad ref")
	}
	r.sendFn(s)
	return nil
}

type processState uint8

const (
	STARTING_STATE processState = iota
	STARTED_STATE
	EXITING_STATE
	EXITED_STATE
)

type Process struct {
	pid       PID
	stateLock sync.RWMutex
	state     processState

	// TODO: Actual Process flags
	// flags flags

	// process management structures; must hold p.stateLock to access or modify these as appropriate
	groupLeader PID
	links       refMap
	monitors    refMap

	// bookkeeping structures; must hold p.stateLock to access or modify these as appropriate
	dirty        bool
	lastGC       time.Time
	gcInterval   time.Duration
	messageSkips []int

	// message passing structures
	// TODO: we should consider a more sophisticated mailbox structure better supporting concurrent receive and send
	// TODO: operations. This will eventually cause contention when many processes are sending messages to the same
	// TODO: process
	mailboxLock sync.Mutex
	mailbox     []Message
}

// TODO: Spawn options
func newProcess[M Message](pid PID, gcInterval time.Duration) *Process {
	debug.Assert(!pid.IsZero(), "newProcess: pid cannot be zero")
	debug.Assert(gcInterval >= 0, "newProcess: gcInterval cannot be negative")
	return &Process{
		pid:         pid,
		groupLeader: PIDZero(),
		lastGC:      time.Now(),
		gcInterval:  gcInterval,
	}
}

func (p *Process) Ref() *Ref {
	return &Ref{pid: p.pid, sendFn: p.handleSignal}
}

func (p *Process) maybeGarbageCollect() {
	if p.shouldGarbageCollect() {
		p.garbageCollect()
	}
}

func (p *Process) shouldGarbageCollect() bool {
	if !p.dirty {
		return false
	}

	if p.lastGC.IsZero() {
		return true
	}
	if p.gcInterval <= 0 {
		return false
	}
	should := time.Since(p.lastGC) >= p.gcInterval
	return should
}

func (p *Process) garbageCollect() {
	if !p.dirty {
		return
	}

	p.links.garbageCollect()
	p.monitors.garbageCollect()

	mailbox := p.mailbox[:0]
	var mi, msi int
	for ; mi < len(p.mailbox); mi++ {
		if msi < len(p.messageSkips) && p.messageSkips[msi] == mi {
			msi++
			continue
		}
		mailbox = append(mailbox, p.mailbox[mi])
	}
	p.mailbox = mailbox
	p.messageSkips = p.messageSkips[:msi]
	p.lastGC = time.Now()
}

// this must always be called while the stateLock is at least read-locked and the mailboxLock is write-locked
func (p *Process) pushMessage(m Message) {
	if p.state == STARTED_STATE {
		p.mailbox = append(p.mailbox, m)
	}
}

// this must always be called while the stateLock is at least read-locked and the mailboxLock is write-locked
func (p *Process) link(re *Ref) {
	p.links.push(re)
}

// this must always be called while the stateLock is at least read-locked and the mailboxLock is write-locked
func (p *Process) unlink(re *Ref) {
	removed := p.links.remove(re)
	p.dirty = p.dirty || removed
}

// this must always be called while the stateLock is at least read-locked and the mailboxLock is write-locked
func (p *Process) exit(e Exit) {
	// TODO: EXIT HANDLING
	panic("TODO: process.exit: not implemented")
}

// this must always be called while the stateLock is at least read-locked and the mailboxLock is write-locked
func (p *Process) monitor(re *Ref) {
	p.monitors.push(re)
}

// this must always be called while the stateLock is at least read-locked and the mailboxLock is write-locked
func (p *Process) deMonitor(re *Ref) {
	removed := p.monitors.remove(re)
	p.dirty = p.dirty || removed
}

// this must always be called while the stateLock is at least read-locked and the mailboxLock is write-locked
func (p *Process) down(down Down) {
	if p.monitors.contains(down.Ref) {
		p.pushMessage(down)
	}
}

// this must always be called while the stateLock is at least read-locked and the mailboxLock is write-locked
func (p *Process) setGroupLeader(pid PID) {
	if pid.IsZero() {
		debug.Throw("process.setGroupLeader: cannot set group leader to zero PID")
	}
	// yes we can be set to ourselves by design, for instance the top supervisor in the
	// [gotp/applicaiton.Application.Start] function will set itself as its own group leader.
	p.groupLeader = pid
}

// this must always be called while the stateLock is at least read-locked and the mailboxLock is write-locked
func (p *Process) aliveRequest(re *Ref) {
	var err error
	if p.state != STARTED_STATE {
		err = errors.New("process not started")
	}
	re.send(aliveReplySignal(p.pid, re, err))
}

// this must always be called while the stateLock is at least read-locked and the mailboxLock is write-locked
func (p *Process) aliveReply(r Reply[error]) {
	p.pushMessage(r)
}

// handleSignal is always called from a functions that has the mailboxLock write-locked
func (p *Process) handleSignal(s signal[Message]) {
	defer p.maybeGarbageCollect()
	p.stateLock.RLock()
	switch p.state {
	case STARTED_STATE:
		switch s._type {
		case LINK_SIGNAL:
			re := debug.AssertType[*Ref](s.message, "process.handleSignal: expected *Ref for LINK_SIGNAL, got %T", s.message)
			p.link(re)
		case UNLINK_SIGNAL:
			re := debug.AssertType[*Ref](s.message, "process.handleSignal: expected *Ref for UNLINK_SIGNAL, got %T", s.message)
			p.unlink(re)
		case EXIT_SIGNAL:
			e := debug.AssertType[Exit](s.message, "process.handleSignal: expected exit for EXIT_SIGNAL, got %T", s.message)
			p.exit(e)
		case MONITOR_SIGNAL:
			re := debug.AssertType[*Ref](s.message, "process.handleSignal: expected *Ref for MONITOR_SIGNAL, got %T", s.message)
			p.monitor(re)
		case DE_MONITOR_SIGNAL:
			re := debug.AssertType[*Ref](s.message, "process.handleSignal: expected *Ref for DE_MONITOR_SIGNAL, got %T", s.message)
			p.deMonitor(re)
		case DOWN_SIGNAL:
			down := debug.AssertType[Down](s.message, "process.handleSignal: expected exit for DOWN_SIGNAL, got %T", s.message)
			p.down(down)
		case GROUP_LEADER_SIGNAL:
			pid := debug.AssertTypeNotZero[PID](s.message, "process.handleSignal: expected PID for GROUP_LEADER_SIGNAL, got %T", s.message)
			p.setGroupLeader(pid)
		case ALIVE_REQUEST_SIGNAL:
			re := debug.AssertType[*Ref](s.message, "process.handleSignal: expected message for ALIVE_REQUEST_SIGNAL, got %T", s.message)
			p.aliveRequest(re)
		case ALIVE_REPLY_SIGNAL:
			err := debug.AssertType[Reply[error]](s.message, "process.handleSignal: expected Reply[error] for ALIVE_REPLY_SIGNAL, got %T", s.message)
			p.aliveReply(err)
		case MESSAGE_SIGNAL:
			p.pushMessage(s.message)
		default:
			p.stateLock.RUnlock()
			// If this is ever hit we should either expect a bad implementation or a new signal type
			// has been added that we don't handle yet.
			debug.Throw("process.handleSignal: unknown signal type: %s", s._type)
			panic("unreachable")
		}
	}
}

// TODO: reassess if reflist might be best because its useful to have ordinality for reverse iteration
type refMap struct {
	count int
	m     map[PID][]*Ref
}

func (rl *refMap) push(re *Ref) {
	debug.Assert(re.IsValid(), "refList.push: cannot push invalid reference")
	v, ok := rl.m[re.pid]
	if !ok {
		v = make([]*Ref, 0, 4)
	}
	v = append(v, re)
	rl.m[re.pid] = v
	rl.count++
}

func (rl *refMap) remove(re *Ref) bool {
	debug.Assert(re.IsValid(), "refList.remove: cannot remove invalid reference")
	vs, ok := rl.m[re.pid]
	if !ok {
		return false
	}
	for i, v := range vs {
		if v == re {
			vs[i] = vs[len(vs)-1]
			vs[len(vs)-1] = nil
			vs = vs[:len(vs)-1]
			rl.m[re.pid] = vs
			rl.count--
			return true
		}
	}
	return false
}

func (rl *refMap) contains(re *Ref) bool {
	debug.Assert(re.IsValid(), "refList.contains: cannot check invalid reference")
	vs, ok := rl.m[re.pid]
	if !ok {
		return false
	}
	for _, v := range vs {
		if v == re {
			return true
		}
	}
	return false
}

func (rl *refMap) garbageCollect() {
	if rl.count == 0 {
		return
	}
	m := rl.m
	if rl.count*4 > len(rl.m) {
		m = make(map[PID][]*Ref, min(16, len(rl.m)*2))
	}
	for pid, refs := range rl.m {
		m[pid] = make([]*Ref, 0, len(refs))
		for _, ref := range refs {
			if ref != nil && ref.IsValid() {
				m[pid] = append(m[pid], ref)
			}
		}
	}
	rl.m = m
}
