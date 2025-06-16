package server

import "github.com/Morgahl/gotp"

// TODO: Placeholder helper to provide easy "optional" Handlers for the Serverable interface. Remove this in favor of
// TODO: breaking each of these out into their own interfaces that are optionally implemented and the Server
// TODO: implementation will only call them if they are implemented otherwise handing them to the HandleAny if
// TODO: implemented and finally just dropped if not implemented.
type DefaultHandlers[Ct gotp.Msg] struct{}

func (DefaultHandlers[Ct]) HandleCall(call any, from gotp.PID) (Response[any], Continue[Ct], error) {
	return NoReply[any](), NoCont[Ct](), nil
}

func (DefaultHandlers[Ct]) HandleCast(cast any) (Continue[Ct], error) {
	return NoCont[Ct](), nil
}

func (DefaultHandlers[Ct]) HandleContinue(cont Ct) (Continue[Ct], error) {
	return NoCont[Ct](), nil
}

func (DefaultHandlers[Ct]) HandleInfo(info gotp.Msg) (Continue[Ct], error) {
	return NoCont[Ct](), nil
}

func (DefaultHandlers[Ct]) Terminate(reason error) error {
	return reason
}
