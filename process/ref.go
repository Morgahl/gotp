package process

import (
	"errors"

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
