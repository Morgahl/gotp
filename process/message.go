package process

type Message interface{}

type Request[M Message] struct {
	From    PID
	Ref     *Ref
	Message M
}

type Reply[M Message] struct {
	From    PID
	Ref     *Ref
	Message M
}
