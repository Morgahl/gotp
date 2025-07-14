package process

import (
	"fmt"
	"log/slog"
)

type Ref struct {
	pid    PID
	sendFn func(s signal[Message])
}

func (r *Ref) IsValid() bool {
	return r.sendFn != nil
}

func (r *Ref) Send(m Message) {
	r.send(messageSignal(no_FLAGS, m))
}

func (r *Ref) send(s signal[Message]) {
	if r.sendFn == nil {
		return
	}
	r.sendFn(s)
}

func (r Ref) String() string {
	if r.IsValid() {
		return fmt.Sprintf("Ref%s", r.pid)
	}

	if r.sendFn != nil {
		return "Ref<opaque>"
	}

	return "Ref<invalid>"
}

func (r Ref) LogValue() slog.Value {
	return slog.StringValue(r.String())
}
