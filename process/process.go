package process

import (
	"errors"
	"log/slog"
	"slices"
	"sync"
	"time"

	"github.com/Morgahl/gotp/debug"
)

type Startable interface {
	Start(opts ...SpawnOpt) (Started, error)
}

type Started interface {
	PID() PID
	Send(Message)
	SendAfter(Message, time.Duration) *time.Timer
	Exit(reason error)
}

type RunFn func(*Process) error

type Process struct {
	pid         PID
	flags       ProcessFlags
	runFn       RunFn
	stateLock   sync.RWMutex
	state       processState
	exitReason  error
	deregHandle func()

	// process management structures; must hold p.stateLock to access or modify these as appropriate
	groupLeader *Ref
	links       refMap
	monitors    refMap

	// bookkeeping structures; must hold p.stateLock to access or modify these as appropriate
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
	p := &Process{
		pid:        nextPID(),
		links:      newRefMap(),
		monitors:   newRefMap(),
		signalChan: make(chan signal[Message], CHANNEL_SIZE),
	}
	for _, opt := range opts {
		opt(p)
	}
	if p.mailbox == nil {
		p.mailbox = make([]Message, 0, MAILBOX_SIZE)
	}
	p.state = STARTING_STATE
	return p
}

func (p *Process) Start() {
	p.stateLock.Lock()
	switch p.state {
	case STARTED_STATE, EXITING_STATE, EXITED_STATE:
		p.stateLock.Unlock()
		debug.Throw("process.Start: cannot start process in state %s", p.state)
	case STARTING_STATE:
		p.state = STARTED_STATE
		p.deregHandle = register(p)
		p.stateLock.Unlock()
		go p.run()
	}
}

func (p *Process) Ref() *Ref {
	return &Ref{
		pid:    p.pid,
		sendFn: p.send,
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

func (p *Process) Exit(reason error) {
	p.send(exitSignal(no_FLAGS, p.pid, p.Ref(), reason))
}

func (p *Process) run() {
	var reason error
	defer func() {
		reason = debug.Recover(recover(), "Process.run", reason)
		p.stateLock.Lock()
		if p.deregHandle != nil {
			p.deregHandle()
			p.deregHandle = nil
		}
		for ref := range p.monitors.refs() {
			ref.send(downSignal(ref.pid, ref, reason))
		}
		for ref := range p.links.refs() {
			ref.send(exitSignal(link_FLAG, ref.pid, ref, reason))
		}
		close(p.signalChan)
		for range p.signalChan {
			// sink the channel to ensure it is closed properly
		}
		p.state = EXITED_STATE
		if p.exitReason == nil {
			p.exitReason = reason
		}
		p.stateLock.Unlock()
	}()
	reason = p.runFn(p)
}

func (p *Process) send(s signal[Message]) {
	p.stateLock.RLock()
	switch p.state {
	case STARTING_STATE, STARTED_STATE:
		p.stateLock.RUnlock()
		p.signalChan <- s
	}
}

func (p *Process) maybeGarbageCollect() {
	if p.shouldGarbageCollect() {
		p.garbageCollect()
	}
}

func (p *Process) shouldGarbageCollect() bool {
	return len(p.messageSkips)*4 > len(p.mailbox)
}

func (p *Process) garbageCollect() {
	p.links.garbageCollect()
	p.monitors.garbageCollect()

	// mailbox GC after here

	switch len(p.messageSkips) {
	case 0:
		return
	case 1:
		skip := p.messageSkips[0]
		mailbox := p.mailbox[:0]
		mailbox = append(mailbox, p.mailbox[:skip]...)
		mailbox = append(mailbox, p.mailbox[skip+1:]...)
		p.mailbox = mailbox
		p.messageSkips = p.messageSkips[:0]
		return
	}

	slices.Sort(p.messageSkips)

	last := -1
	mailbox := p.mailbox[:0]
	for msi := 0; msi < len(p.messageSkips); msi++ {
		skip := p.messageSkips[msi]
		mailbox = append(mailbox, p.mailbox[last+1:skip]...)
		last = skip
	}
	if last+1 < len(p.mailbox) {
		mailbox = append(mailbox, p.mailbox[last+1:]...)
	}

	p.mailbox = mailbox
	p.messageSkips = p.messageSkips[:0]
}

// this must always be called while the stateLock is at least read-locked and the mailboxLock is write-locked
func (p *Process) pushMessage(m Message) {
	if p.state == STARTED_STATE {
		p.mailbox = append(p.mailbox, m)
	}
}

// this must always be called while the stateLock is at least read-locked and the mailboxLock is write-locked
func (p *Process) linkRequest(re RequestMsg[*Ref]) {
	p.links.push(re.From, re.Message)
	ref := p.Ref()
	re.Ref.send(linkReplySignal(ReplyMsg[*Ref]{
		From:    p.pid,
		Ref:     ref,
		Message: re.Message,
	}))
}

func (p *Process) linkReply(re ReplyMsg[*Ref]) {
	p.links.push(re.From, re.Message)
}

// this must always be called while the stateLock is at least read-locked and the mailboxLock is write-locked
func (p *Process) unlink(re RequestMsg[*Ref]) {
	p.links.remove(re.From, re.Message)
}

// this must always be called while the stateLock is at least read-locked and the mailboxLock is write-locked
func (p *Process) monitor(re RequestMsg[*Ref]) {
	p.monitors.push(re.From, re.Message)
}

// this must always be called while the stateLock is at least read-locked and the mailboxLock is write-locked
func (p *Process) deMonitor(re RequestMsg[*Ref]) {
	p.monitors.remove(re.From, re.Message)
}

// this must always be called while the stateLock is at least read-locked and the mailboxLock is write-locked
func (p *Process) handleExitSignal(f signalFlags, e exitSig) {
	linked := f.IsLink()
	linkFound := p.links.contains(e.PID, e.Receiver)
	trappingExits := p.flags.IsTrapExit()
	samePid := e.Receiver != nil && e.Receiver.pid == e.PID

	// Silently drop the exit signal if:
	// - it is a link signal and the receiver is not linked to the sender
	// - it is a normal exit and the process is not trapping exits and the sender
	//   is not the same as the receiver
	if (linked && !linkFound) ||
		(e.Reason == KILL && !trappingExits && !samePid) {
		slog.Debug("process.handleExitSignal: silently dropping exit signal", slog.Any("pid", p.pid), slog.Any("exit", e))
		return
	}

	// Terminate the receiving process if:
	// - it is not a link signal and the exit is `kill`; the receiver is killed with the `killed` reason
	// - the process is not trapping exits, the exit reason is something other than `normal`
	// - the exit reason is `normal` and the sender is the same as the receiver and the link flag is not set
	if !linked && e.Reason == KILL {
		slog.Debug("process.handleExitSignal: terminating process with KILL reason", slog.Any("pid", p.pid), slog.Any("exit", e))
		p.stateLock.RUnlock()
		p.stateLock.Lock()
		p.state = EXITING_STATE
		p.exitReason = KILLED
		p.stateLock.Unlock()
		p.stateLock.RLock()
		return
	} else if !trappingExits && e.Reason != NORMAL {
		slog.Debug("process.handleExitSignal: terminating process with exit reason", slog.Any("pid", p.pid), slog.Any("exit", e))
		p.stateLock.RUnlock()
		p.stateLock.Lock()
		p.state = EXITING_STATE
		p.exitReason = e.Reason
		p.stateLock.Unlock()
		p.stateLock.RLock()
		return
	} else if e.Reason == NORMAL && samePid && !linked {
		slog.Debug("process.handleExitSignal: terminating process with normal exit reason", slog.Any("pid", p.pid), slog.Any("exit", e))
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
		slog.Debug("process.handleExitSignal: pushing exit message to mailbox", slog.Any("pid", p.pid), slog.Any("exit", e))
		p.pushMessage(e.ToExit())
	}
}

// this must always be called while the stateLock is at least read-locked and the mailboxLock is write-locked
func (p *Process) down(down DownMsg) {
	if p.monitors.contains(down.From, down.Ref) {
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
func (p *Process) aliveRequest(req RequestMsg[Message]) {
	var err error
	if p.state != STARTED_STATE {
		err = errors.New("process not started")
	}
	sig := aliveReplySignal(req, p.Ref(), err)
	if req.Ref != nil {
		req.Ref.send(sig)
	} else {
		sendPID(req.From, sig)
	}
}

// this must always be called while the stateLock is at least read-locked and the mailboxLock is write-locked
func (p *Process) aliveReply(r ReplyMsg[error]) {
	p.pushMessage(r)
}

// handleSignal is always called from a functions that has the mailboxLock write-locked as well as the stateLock read-locked.
func (p *Process) handleSignal(s signal[Message]) {
	slog.Debug("process.handleSignal", slog.Any("pid", p.pid), slog.Any("signal", s))
	switch p.state {
	case STARTED_STATE:
		switch s._type {
		case LINK_SIGNAL:
			if s.flags.IsRequest() {
				re := debug.AssertType[RequestMsg[*Ref]](s.message, "process.handleSignal: expected *Ref for LINK_SIGNAL, got %T", s.message)
				p.linkRequest(re)
			} else if s.flags.IsReply() {
				re := debug.AssertType[ReplyMsg[*Ref]](s.message, "process.handleSignal: expected *Ref for LINK_REPLY_SIGNAL, got %T", s.message)
				p.linkReply(re)
			} else {
				debug.Throw("process.handleSignal: unexpected flags for LINK_SIGNAL: %s", s.flags)
			}
		case UNLINK_SIGNAL:
			re := debug.AssertType[RequestMsg[*Ref]](s.message, "process.handleSignal: expected *Ref for UNLINK_SIGNAL, got %T", s.message)
			p.unlink(re)
		case EXIT_SIGNAL:
			e := debug.AssertType[exitSig](s.message, "process.handleSignal: expected exit for EXIT_SIGNAL, got %T", s.message)
			p.handleExitSignal(s.flags, e)
		case MONITOR_SIGNAL:
			re := debug.AssertType[RequestMsg[*Ref]](s.message, "process.handleSignal: expected *Ref for MONITOR_SIGNAL, got %T", s.message)
			p.monitor(re)
		case DE_MONITOR_SIGNAL:
			re := debug.AssertType[RequestMsg[*Ref]](s.message, "process.handleSignal: expected *Ref for DE_MONITOR_SIGNAL, got %T", s.message)
			p.deMonitor(re)
		case DOWN_SIGNAL:
			down := debug.AssertType[DownMsg](s.message, "process.handleSignal: expected exit for DOWN_SIGNAL, got %T", s.message)
			p.down(down)
		case GROUP_LEADER_SIGNAL:
			re := debug.AssertType[*Ref](s.message, "process.handleSignal: expected *Ref for GROUP_LEADER_SIGNAL, got %T", s.message)
			p.setGroupLeader(re)
		case ALIVE_REQUEST_SIGNAL:
			req := debug.AssertType[RequestMsg[Message]](s.message, "process.handleSignal: expected message for ALIVE_REQUEST_SIGNAL, got %T", s.message)
			p.aliveRequest(req)
		case ALIVE_REPLY_SIGNAL:
			err := debug.AssertType[ReplyMsg[error]](s.message, "process.handleSignal: expected Reply[error] for ALIVE_REPLY_SIGNAL, got %T", s.message)
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

func newRefMap() refMap { return refMap{m: make(map[PID][]*Ref, 16)} }

func (rl *refMap) push(id PID, re *Ref) {
	debug.RefuteFunc(id.IsZero, "refList.push: cannot push zero PID")
	debug.AssertFunc(re.IsValid, "refList.push: cannot push invalid reference")
	v, ok := rl.m[id]
	if !ok {
		v = make([]*Ref, 0, 4)
	}
	v = append(v, re)
	rl.m[re.pid] = v
	rl.count++
}

func (rl *refMap) remove(id PID, re *Ref) bool {
	debug.RefuteFunc(id.IsZero, "refList.remove: cannot remove zero PID")
	debug.AssertFunc(re.IsValid, "refList.remove: cannot remove invalid reference")
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

func (rl *refMap) contains(id PID, re *Ref) bool {
	debug.RefuteFunc(id.IsZero, "refList.contains: cannot check zero PID")
	debug.AssertFunc(re.IsValid, "refList.contains: cannot check invalid reference")
	vs, ok := rl.m[id]
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
	if rl.count*4 > len(rl.m) {
		m := make(map[PID][]*Ref, min(16, len(rl.m)*2))
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
