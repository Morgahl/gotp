package process

import (
	"context"
	"errors"
	"slices"
	"sync"
	"time"

	"github.com/Morgahl/gotp"
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
	pid             PID
	name            gotp.Atom
	flags           ProcessFlags
	runFn           RunFn
	state           processState
	exitReason      error
	context         context.Context
	contextCancel   context.CancelCauseFunc
	deregPidHandle  func()
	deregNameHandle func()

	// message passing structures; must hold p.mailboxMu to access or modify these as appropriate
	mailboxMu  sync.Mutex
	mailbox    []Message
	signalChan chan signal[Message]

	// bookkeeping structures; must hold p.mailboxMu to access or modify these as appropriate
	messageSkips []int

	// process management structures; must hold p.mailboxMu to access or modify these as appropriate
	groupLeader Ref
	links       refMap
	monitors    refMap
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
	if p.context == nil {
		p.context, p.contextCancel = context.WithCancelCause(context.Background())
	}
	p.context = context.WithValue(p.context, gotp.Atom("pid"), p.pid)
	if p.name != "" {
		p.context = context.WithValue(p.context, gotp.Atom("name"), p.name)
	}
	return p
}

func (p *Process) Start() {
	switch p.state {
	case STARTED_STATE, EXITING_STATE, EXITED_STATE:
		debug.Throw("process.Start: cannot start process in state %s", p.state)
	case STARTING_STATE:
		p.state = STARTED_STATE
		p.deregPidHandle = registerPID(p)
		if p.name != "" {
			p.deregNameHandle = registerNamed(p.name, p.Ref())
		}
		go p.run()
	}
}

func (p *Process) Ref() Ref {
	return newRef(p)
}

func (p *Process) Context() context.Context {
	return p.context
}

func (p *Process) PID() PID {
	return p.pid
}

func (p *Process) Name() gotp.Atom {
	return p.name
}

func (p *Process) UpdateFlags(fn func(ProcessFlags) ProcessFlags) {
	p.flags = fn(p.flags)
}

func (p *Process) Exit(reason error) {
	p.send(exitSignal(no_FLAGS, p.pid, p.Ref(), reason))
}

func (p *Process) run() {
	defer func() {
		p.exitReason = debug.Recover(recover(), "Process.run", p.exitReason)
		p.mailboxMu.Lock()
		if p.deregNameHandle != nil {
			p.deregNameHandle()
			p.deregNameHandle = nil
		}
		if p.deregPidHandle != nil {
			p.deregPidHandle()
			p.deregPidHandle = nil
		}
		for ref := range p.monitors.refs() {
			ref.send(downSignal(p.pid, ref, p.exitReason))
		}
		for ref := range p.links.refs() {
			ref.send(exitSignal(link_FLAG, p.pid, ref, p.exitReason))
		}
		close(p.signalChan)
		p.signalChan = nil
		p.state = EXITED_STATE
		p.mailboxMu.Unlock()
	}()
	if err := p.runFn(p); err != nil {
		p.exitReason = errors.Join(p.exitReason, err)
	}
}

func (p *Process) send(s signal[Message]) {
	defer func() { recover() }()
	p.signalChan <- s
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
func (p *Process) linkRequest(re RequestMsg[Ref]) {
	p.links.push(re.From, re.Message)
	ref := p.Ref()
	re.Ref.send(linkReplySignal(ReplyMsg[Ref]{
		From:    p.pid,
		Ref:     ref,
		Message: re.Message,
	}))
}

func (p *Process) linkReply(re ReplyMsg[Ref]) {
	p.links.push(re.From, re.Message)
}

// this must always be called while the stateLock is at least read-locked and the mailboxLock is write-locked
func (p *Process) unlink(re RequestMsg[Ref]) {
	p.links.remove(re.From, re.Message)
}

// this must always be called while the stateLock is at least read-locked and the mailboxLock is write-locked
func (p *Process) monitor(re RequestMsg[Ref]) {
	p.monitors.push(re.From, re.Message)
}

// this must always be called while the stateLock is at least read-locked and the mailboxLock is write-locked
func (p *Process) deMonitor(re RequestMsg[Ref]) {
	p.monitors.remove(re.From, re.Message)
}

// this must always be called while the stateLock is at least read-locked and the mailboxLock is write-locked
func (p *Process) handleExitSignal(f signalFlags, e exitSig) {
	linked := f.IsLink()
	linkFound := p.links.contains(e.PID, e.Ref)
	trappingExits := p.flags.IsTrapExit()
	samePid := e.Ref.pid == e.PID

	// clean up the link if it was a link signal and the receiver is linked to the sender
	if linked && linkFound {
		p.links.remove(e.PID, e.Ref)
	}
	// Silently drop the exit signal if:
	// - it is a link signal and the receiver is not linked to the sender
	// - it is a normal exit and the process is not trapping exits and the sender
	//   is not the same as the receiver
	if (linked && !linkFound) ||
		(errors.Is(e.Reason, KILL) && !trappingExits && !samePid) {
		return
	}

	// Terminate the receiving process if:
	// - it is not a link signal and the exit is `kill`; the receiver is killed with the `killed` reason
	// - the process is not trapping exits, the exit reason is something other than `normal`
	// - the exit reason is `normal` and the sender is the same as the receiver and the link flag is not set
	if !linked && errors.Is(e.Reason, KILL) {
		p.state = EXITING_STATE
		p.exitReason = KILLED
		return
	} else if !trappingExits && !errors.Is(e.Reason, NORMAL) {
		p.state = EXITING_STATE
		p.exitReason = e.Reason
		return
	} else if errors.Is(e.Reason, NORMAL) && samePid && !linked {
		p.state = EXITING_STATE
		p.exitReason = NORMAL
		return
	}

	// The exit is converted to e message and pushed to the mailbox if the process is trapping exits and the link flag
	// is:
	// - not set, and the exit reason is not `kill`
	// - set, the receiver is linked to the sender
	if trappingExits ||
		(!linked && !errors.Is(e.Reason, KILL)) ||
		(linked && linkFound) {
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
func (p *Process) setGroupLeader(re Ref) {
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
	if req.Ref.IsValid() {
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
	switch p.state {
	case STARTED_STATE:
		switch s._type {
		case LINK_SIGNAL:
			if s.flags.IsRequest() {
				re := debug.AssertType[RequestMsg[Ref]](s.message, "process.handleSignal: expected Ref for LINK_SIGNAL, got %T", s.message)
				p.linkRequest(re)
			} else if s.flags.IsReply() {
				re := debug.AssertType[ReplyMsg[Ref]](s.message, "process.handleSignal: expected Ref for LINK_REPLY_SIGNAL, got %T", s.message)
				p.linkReply(re)
			} else {
				debug.Throw("process.handleSignal: unexpected flags for LINK_SIGNAL: %s", s.flags)
			}
		case UNLINK_SIGNAL:
			re := debug.AssertType[RequestMsg[Ref]](s.message, "process.handleSignal: expected Ref for UNLINK_SIGNAL, got %T", s.message)
			p.unlink(re)
		case EXIT_SIGNAL:
			e := debug.AssertType[exitSig](s.message, "process.handleSignal: expected exit for EXIT_SIGNAL, got %T", s.message)
			p.handleExitSignal(s.flags, e)
		case MONITOR_SIGNAL:
			re := debug.AssertType[RequestMsg[Ref]](s.message, "process.handleSignal: expected Ref for MONITOR_SIGNAL, got %T", s.message)
			p.monitor(re)
		case DE_MONITOR_SIGNAL:
			re := debug.AssertType[RequestMsg[Ref]](s.message, "process.handleSignal: expected Ref for DE_MONITOR_SIGNAL, got %T", s.message)
			p.deMonitor(re)
		case DOWN_SIGNAL:
			down := debug.AssertType[DownMsg](s.message, "process.handleSignal: expected exit for DOWN_SIGNAL, got %T", s.message)
			p.down(down)
		case GROUP_LEADER_SIGNAL:
			re := debug.AssertType[Ref](s.message, "process.handleSignal: expected Ref for GROUP_LEADER_SIGNAL, got %T", s.message)
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
	m     map[PID][]Ref
}

func newRefMap() refMap { return refMap{m: make(map[PID][]Ref, 16)} }

func (rl *refMap) push(id PID, re Ref) {
	debug.RefuteFunc(id.IsZero, "refList.push: cannot push zero PID")
	debug.AssertFunc(re.IsValid, "refList.push: cannot push invalid reference")
	v, ok := rl.m[id]
	if !ok {
		v = make([]Ref, 0, 4)
	}
	v = append(v, re)
	rl.m[id] = v
	rl.count++
}

func (rl *refMap) remove(id PID, re Ref) bool {
	debug.RefuteFunc(id.IsZero, "refList.remove: cannot remove zero PID")
	debug.AssertFunc(re.IsValid, "refList.remove: cannot remove invalid reference")
	vs, ok := rl.m[id]
	if !ok {
		return false
	}
	for i, v := range vs {
		if v == re {
			vs[i] = vs[len(vs)-1]
			vs[len(vs)-1] = Ref{}
			vs = vs[:len(vs)-1]
			rl.m[id] = vs
			rl.count--
			return true
		}
	}
	return false
}

func (rl *refMap) contains(id PID, re Ref) bool {
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
		m := make(map[PID][]Ref, min(16, len(rl.m)*2))
		for pid, refs := range rl.m {
			m[pid] = make([]Ref, 0, len(refs))
			for _, ref := range refs {
				if ref.IsValid() {
					m[pid] = append(m[pid], ref)
				}
			}
		}
		rl.m = m
	}
}

func (rl *refMap) refs() func(func(Ref) bool) {
	return func(yield func(Ref) bool) {
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
