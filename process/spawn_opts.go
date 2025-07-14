package process

import (
	"github.com/Morgahl/gotp/debug"
)

// TODO: rethink the *Process passing here we may want this to just be a builder struct instead for safety
type SpawnOpt func(*Process)

func Linked(pp *Process) SpawnOpt {
	ref := pp.Ref()
	s := linkRequestSignal(RequestMsg[*Ref]{
		From:    pp.PID(),
		Ref:     ref,
		Message: ref,
	})
	// TODO: A send like this might deadlock during building as the process is not receiving yet
	return func(p *Process) { p.send(s) }
}

// TODO: rethink the *Process passing here we may want this to just be a builder struct instead for safety
func Monitored(p *Process) SpawnOpt {
	ref := p.Ref()
	s := monitorSignal(RequestMsg[*Ref]{
		From:    p.PID(),
		Ref:     ref,
		Message: ref,
	})
	// TODO: A send like this might deadlock during building as the process is not receiving yet
	return func(p *Process) { p.send(s) }
}

func GroupLeader(leader *Ref) SpawnOpt {
	debug.Assert(leader.IsValid(), "group leader must be a valid reference")
	return func(p *Process) {
		p.groupLeader = leader
	}
}

func MailboxSize(size int) SpawnOpt {
	if size < 0 {
		size = MAILBOX_SIZE
	}
	return func(p *Process) {
		if p.mailbox == nil {
			p.mailbox = make([]Message, 0, size)
		} else {
			capacity := cap(p.mailbox)
			if capacity < size {
				p.mailbox = make([]Message, 0, size)
			}
		}
	}
}

func ChannelSize(size int) SpawnOpt {
	if size < 0 {
		size = CHANNEL_SIZE
	}
	return func(p *Process) {
		if cap(p.signalChan) < size {
			oldChan := p.signalChan
			p.signalChan = make(chan signal[Message], size)
			if len(oldChan) > 0 {
				close(oldChan)
				for msg := range oldChan {
					p.signalChan <- msg
				}
			}
		}
	}
}

func InheritFrom(p *Process) SpawnOpt {
	groupLeader := p.groupLeader
	return func(p *Process) {
		p.groupLeader = groupLeader
	}
}
