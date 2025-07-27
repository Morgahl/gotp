package process

import (
	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/assert"
)

type SpawnOpt func(*process)

func Link(ref Ref) SpawnOpt {
	s := linkRequestSignal(RequestMsg[Ref]{
		From:    ref.pid,
		Ref:     ref,
		Message: ref,
	})
	// TODO: A send like this might deadlock during building as the process is not receiving yet
	return func(p *process) {
		assert.Equal(p.state, STARTING_STATE, "process must be in STARTING_STATE for spawn options to be applied")
		p.send(s)
	}
}

// TODO: rethink the *process passing here we may want this to just be a builder struct instead for safety
func Monitored(ref Ref) SpawnOpt {
	s := monitorSignal(RequestMsg[Ref]{
		From:    ref.pid,
		Ref:     ref,
		Message: ref,
	})
	// TODO: A send like this might deadlock during building as the process is not receiving yet
	return func(p *process) {
		assert.Equal(p.state, STARTING_STATE, "process must be in STARTING_STATE for spawn options to be applied")
		p.send(s)
	}
}

func MailboxSize(size int) SpawnOpt {
	if size < 0 {
		size = MAILBOX_SIZE
	}
	return func(p *process) {
		assert.Equal(p.state, STARTING_STATE, "process must be in STARTING_STATE for spawn options to be applied")
		if p.mailbox == nil {
			p.mailbox = make([]gotp.Term, 0, size)
		} else {
			capacity := cap(p.mailbox)
			if capacity < size {
				p.mailbox = make([]gotp.Term, 0, size)
			}
		}
	}
}

func ChannelSize(size int) SpawnOpt {
	if size < 0 {
		size = CHANNEL_SIZE
	}
	return func(p *process) {
		assert.Equal(p.state, STARTING_STATE, "process must be in STARTING_STATE for spawn options to be applied")
		if cap(p.signalChan) < size {
			oldChan := p.signalChan
			p.signalChan = make(chan signal[gotp.Term], size)
			if len(oldChan) > 0 {
				close(oldChan)
				for msg := range oldChan {
					p.signalChan <- msg
				}
			}
		}
	}
}

func Named(name gotp.Atom) SpawnOpt {
	return func(p *process) {
		assert.Equal(p.state, STARTING_STATE, "process must be in STARTING_STATE for spawn options to be applied")
		assert.Equal(p.name, "", "process name must be empty when setting it")
		p.name = name
	}
}
