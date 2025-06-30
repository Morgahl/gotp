package process

type Ref struct {
	pid    PID
	sendFn func(s signal[Message])
}

func (r *Ref) IsValid() bool {
	return r.sendFn != nil
}

func (r *Ref) Send(m Message) {
	r.send(messageSignal(NO_FLAGS, m))
}

func (r *Ref) send(s signal[Message]) {
	if r.sendFn == nil {
		return
	}
	r.sendFn(s)
}
