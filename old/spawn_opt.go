package gotp

type SpawnOpt func(*Process)

func Link(pid PID) SpawnOpt {
	return func(p *Process) {
		p.linked = pid
	}
}

func MailboxSize(size int) SpawnOpt {
	if size < 0 {
		size = MAILBOX_SIZE
	}
	return func(p *Process) {
		p.mbSize = size
	}
}
