package process

import "time"

type SpawnOpt func(*Process)

func Link(link *Ref) SpawnOpt {
	return func(p *Process) {
		p.links.push(link)
		link.send(linkSignal(p.Ref()))
	}
}

func Linked(pp *Process) SpawnOpt {
	return Link(pp.Ref())
}

func Monitor(monitor *Ref) SpawnOpt {
	return func(p *Process) {
		monitor.send(monitorSignal(p.Ref()))
	}
}

func Monitored(pp *Process) SpawnOpt {
	return Monitor(pp.Ref())
}

func GroupLeader(leader *Ref) SpawnOpt {
	return func(p *Process) {
		p.groupLeader = leader
	}
}

func GCInterval(interval time.Duration) SpawnOpt {
	if interval < 0 {
		interval = time.Second
	}
	return func(p *Process) {
		p.gcInterval = interval
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
		if p.signalChan == nil {
			p.signalChan = make(chan signal[Message], size)
		} else {
			capacity := cap(p.signalChan)
			if capacity < size {
				p.signalChan = make(chan signal[Message], size)
			}
		}
	}
}

func InheritFrom(pp *Process) SpawnOpt {
	flags := pp.flags
	groupLeader := pp.groupLeader
	return func(p *Process) {
		p.flags = flags
		p.groupLeader = groupLeader
	}
}
