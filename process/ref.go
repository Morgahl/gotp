package process

import (
	"fmt"
	"log/slog"
	"weak"
)

type Ref struct {
	pid     PID
	procRef weak.Pointer[Process]
}

func newRef(p *Process) Ref {
	return Ref{
		pid:     p.pid,
		procRef: weak.Make(p),
	}
}

func (r Ref) IsValid() bool {
	return r.procRef.Value() != nil
}

func (r Ref) Send(m Message) {
	r.send(messageSignal(no_FLAGS, m))
}

func (r Ref) send(s signal[Message]) {
	if proc := r.procRef.Value(); proc != nil {
		proc.send(s)
	}
}

func (r Ref) String() string {
	if !r.IsValid() {
		return "Ref<invalid>"
	}
	if !r.pid.IsZero() {
		return fmt.Sprintf("Ref%s", r.pid)
	}
	return "Ref<opaque>"
}

func (r Ref) LogValue() slog.Value {
	return slog.StringValue(r.String())
}
