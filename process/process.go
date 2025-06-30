package process

import (
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/Morgahl/gotp/debug"
)

type RunFn func(*Process) error

type Process struct {
	pid         PID
	flags       ProcessFlags
	runFn       RunFn
	stateLock   sync.RWMutex
	state       processState
	exitReason  fmt.Stringer
	deregHandle func()

	// process management structures; must hold p.stateLock to access or modify these as appropriate
	groupLeader *Ref
	links       refMap
	monitors    refMap

	// bookkeeping structures; must hold p.stateLock to access or modify these as appropriate
	dirty        bool
	lastGC       time.Time
	gcInterval   time.Duration
	messageSkips []int

	// message passing structures
	mailboxMu  sync.RWMutex
	mailbox    []Message
	signalChan chan signal[Message]
}

func Spawn(fn RunFn, opts ...SpawnOpt) *Process {
	p := build(opts)
	p.runFn = fn
	return p
}

func SpawnLink(fn RunFn, linked *Process, opts ...SpawnOpt) *Process {
	linkOpts := []SpawnOpt{
		Linked(linked),
		InheritFrom(linked),
	}
	return Spawn(fn, append(linkOpts, opts...)...)
}

func SpawnMonitor(fn RunFn, monitor *Process, opts ...SpawnOpt) *Process {
	monitorOpts := []SpawnOpt{
		Monitored(monitor),
		InheritFrom(monitor),
	}
	return Spawn(fn, append(monitorOpts, opts...)...)
}

func build(opts []SpawnOpt) *Process {
	pid := nextPID()
	p := &Process{
		gcInterval: time.Second,
	}
	for _, opt := range opts {
		opt(p)
	}
	if p.signalChan == nil {
		p.signalChan = make(chan signal[Message], CHANNEL_SIZE)
	}
	if p.mailbox == nil {
		p.mailbox = make([]Message, 0, MAILBOX_SIZE)
	}
	if p.groupLeader == nil {
		p.groupLeader = p.Ref()
	}
	p.pid = pid
	p.state = STARTED_STATE
	return p
}

func (p *Process) Start() {
	p.stateLock.Lock()
	defer p.stateLock.Unlock()
	switch p.state {
	case STARTED_STATE, EXITING_STATE, EXITED_STATE:
		debug.Throw("process.Start: cannot start process in state %s", p.state)
	case STARTING_STATE:
		p.state = STARTED_STATE
		p.deregHandle = register(p)
		go p.run()
	}
}

func (p *Process) Ref() *Ref {
	return &Ref{
		pid:    p.pid,
		sendFn: func(s signal[Message]) { p.send(s) },
	}
}

func (p *Process) PID() PID {
	return p.pid
}

func (p *Process) UpdateFlags(fn func(ProcessFlags) ProcessFlags) {
	p.stateLock.Lock()
	defer p.stateLock.Unlock()
	p.flags = fn(p.flags)
}

func (p *Process) run() {
	var reason error
	defer func() {
		reason = debug.Recover(recover(), "Process.run", reason)
		p.stateLock.Lock()
		defer p.stateLock.Unlock()
		if p.deregHandle != nil {
			p.deregHandle()
			p.deregHandle = nil
		}
		for ref := range p.monitors.refs() {
			ref.send(downSignal(p.PID(), ref, reason))
		}
		for ref := range p.links.refs() {
			ref.send(exitSignal(link_FLAG, p.PID(), ref, Reason{reason}))
		}
	}()
	reason = p.runFn(p)
}

func (p *Process) send(s signal[Message]) {
	p.stateLock.RLock()
	switch p.state {
	case STARTING_STATE, STARTED_STATE:
		p.signalChan <- s
	}
	p.stateLock.RUnlock()
}

func (p *Process) maybeGarbageCollect() {
	if p.shouldGarbageCollect() {
		p.stateLock.RUnlock()
		p.stateLock.Lock()
		p.garbageCollect()
		p.stateLock.Unlock()
		p.stateLock.RLock()
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
	slices.Sort(p.messageSkips)
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
func (p *Process) handleExitSignal(f signalFlags, e exit) {
	linked := f.IsLink()
	linkFound := p.links.contains(e.Receiver)
	trappingExits := p.flags.IsTrapExit()
	samePid := e.Sender == e.Receiver.pid
	// Silently drop the exit signal if:
	// - it is a link signal and the receiver is not linked to the sender
	// - it is a normal exit and the process is not trapping exits and the sender
	//   is not the same as the receiver
	if (linked && !linkFound) ||
		(e.Reason == KILL && !trappingExits && !samePid) {
		return
	}

	// Terminate the receiving process if:
	// - it is not a link signal and the exit is `kill`; the receiver is killed with the `killed` reason
	// - the process is not trapping exits, the exit reason is something other than `normal`
	// - the exit reason is `normal` and the sender is the same as the receiver and the link flag is not set
	if !linked && e.Reason == KILL {
		p.stateLock.RUnlock()
		p.stateLock.Lock()
		p.state = EXITING_STATE
		p.exitReason = KILLED
		p.stateLock.Unlock()
		p.stateLock.RLock()
		return
	} else if !trappingExits && e.Reason != NORMAL {
		p.stateLock.RUnlock()
		p.stateLock.Lock()
		p.state = EXITING_STATE
		p.exitReason = e.Reason
		p.stateLock.Unlock()
		p.stateLock.RLock()
		return
	} else if e.Reason == NORMAL && samePid && !linked {
		p.stateLock.RUnlock()
		p.stateLock.Lock()
		p.state = EXITING_STATE
		p.exitReason = NORMAL
		p.stateLock.Unlock()
		p.stateLock.RLock()
		return
	}

	// The exit is converted to e message and pushed to the mailbox if the process is trapping exits and the link flag
	// is:
	// - not set, and the exit reason is not `kill`
	// - set, the receiver is linked to the sender
	if trappingExits ||
		(!linked && e.Reason != KILL) ||
		(linked && linkFound) {
		p.pushMessage(e.ToExit())
	}
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
func (p *Process) setGroupLeader(re *Ref) {
	// yes we can be set to ourselves by design, for instance the top supervisor in the
	// [gotp/applicaiton.Application.Start] function will set itself as its own group leader.
	p.groupLeader = re
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

// handleSignal is always called from a functions that has the mailboxLock write-locked as well as the stateLock read-locked.
func (p *Process) handleSignal(s signal[Message]) {
	p.stateLock.RLock()
	defer p.stateLock.RUnlock()
	defer p.maybeGarbageCollect()
	switch p.state {
	case STARTED_STATE:
		switch s._type {
		case LINK_SIGNAL:
			re := debug.AssertTypeNotZero[*Ref](s.message, "process.handleSignal: expected *Ref for LINK_SIGNAL, got %T", s.message)
			p.link(re)
		case UNLINK_SIGNAL:
			re := debug.AssertTypeNotZero[*Ref](s.message, "process.handleSignal: expected *Ref for UNLINK_SIGNAL, got %T", s.message)
			p.unlink(re)
		case EXIT_SIGNAL:
			e := debug.AssertTypeNotZero[exit](s.message, "process.handleSignal: expected exit for EXIT_SIGNAL, got %T", s.message)
			p.handleExitSignal(s.flags, e)
		case MONITOR_SIGNAL:
			re := debug.AssertTypeNotZero[*Ref](s.message, "process.handleSignal: expected *Ref for MONITOR_SIGNAL, got %T", s.message)
			p.monitor(re)
		case DE_MONITOR_SIGNAL:
			re := debug.AssertTypeNotZero[*Ref](s.message, "process.handleSignal: expected *Ref for DE_MONITOR_SIGNAL, got %T", s.message)
			p.deMonitor(re)
		case DOWN_SIGNAL:
			down := debug.AssertTypeNotZero[Down](s.message, "process.handleSignal: expected exit for DOWN_SIGNAL, got %T", s.message)
			p.down(down)
		case GROUP_LEADER_SIGNAL:
			re := debug.AssertTypeNotZero[*Ref](s.message, "process.handleSignal: expected *Ref for GROUP_LEADER_SIGNAL, got %T", s.message)
			p.setGroupLeader(re)
		case ALIVE_REQUEST_SIGNAL:
			re := debug.AssertTypeNotZero[*Ref](s.message, "process.handleSignal: expected message for ALIVE_REQUEST_SIGNAL, got %T", s.message)
			p.aliveRequest(re)
		case ALIVE_REPLY_SIGNAL:
			err := debug.AssertTypeNotZero[Reply[error]](s.message, "process.handleSignal: expected Reply[error] for ALIVE_REPLY_SIGNAL, got %T", s.message)
			p.aliveReply(err)
		case MESSAGE_SIGNAL:
			p.pushMessage(s.message)
		default:
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

func (rl *refMap) refs() func(func(*Ref) bool) {
	return func(yield func(*Ref) bool) {
		for _, refs := range rl.m {
			for _, ref := range refs {
				if !yield(ref) {
					return
				}
			}
		}
	}
}

type processState uint8

const (
	STARTING_STATE processState = iota
	STARTED_STATE
	EXITING_STATE
	EXITED_STATE
)

func (s processState) String() string {
	switch s {
	case STARTING_STATE:
		return "STARTING"
	case STARTED_STATE:
		return "STARTED"
	case EXITING_STATE:
		return "EXITING"
	case EXITED_STATE:
		return "EXITED"
	default:
		debug.Throw("processState.String: unknown process state: %d", s)
		return "UNKNOWN"
	}
}
