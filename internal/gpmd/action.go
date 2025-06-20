package gpmd

import (
	"encoding/json"
	"fmt"
)

type ActionType uint8

const (
	ActionUnknown ActionType = iota
	ActionRegister
	ActionUnregister
	ActionHeartbeat
	ActionNames
	ActionKill
)

func (action ActionType) String() string {
	switch action {
	case ActionRegister:
		return "REGISTER"
	case ActionUnregister:
		return "UNREGISTER"
	case ActionHeartbeat:
		return "HEARTBEAT"
	case ActionNames:
		return "NAMES"
	case ActionKill:
		return "KILL"
	default:
		return "UNKNOWN"
	}
}

type ActionRequest struct {
	ActionType ActionType `json:"actionType"`
	Node       Node       `json:"node,omitempty"`
}

func (ar ActionRequest) String() string {
	return fmt.Sprintf("ActionRequest{ActionType:%s, Node:%s}", ar.ActionType, ar.Node)
}

type JSONable interface {
	json.Marshaler
	json.Unmarshaler
}

type ActionResult[R JSONable] struct {
	ActionType ActionType `json:"actionType"`
	Result     Result[R]
}

func SuccessResult[R JSONable](actionType ActionType, result R) ActionResult[R] {
	return ActionResult[R]{
		ActionType: actionType,
		Result:     Result[R]{OK: result},
	}
}

func ErrorResult[R JSONable](actionType ActionType, err error) ActionResult[R] {
	return ActionResult[R]{
		ActionType: actionType,
		Result:     Result[R]{Error: err},
	}
}

func (ar ActionResult[R]) String() string {
	return fmt.Sprintf("ActionResult{ActionType:%s, Result:%v}", ar.ActionType, ar.Result)
}

type None struct{}

func (None) MarshalJSON() ([]byte, error) {
	return []byte("null"), nil
}

func (None) UnmarshalJSON([]byte) error {
	return nil
}

type Result[R JSONable] struct {
	OK    R     `json:"ok,omitempty"`
	Error error `json:"error,omitempty"`
}

func (r Result[R]) String() string {
	if r.Error != nil {
		return fmt.Sprintf("Result{Error:%s}", r.Error)
	}

	if ok, stringer := any(r.OK).(fmt.Stringer); stringer {
		return fmt.Sprintf("Result{OK:%s}", ok)
	}

	return fmt.Sprintf("Result{OK:%+v}", r.OK)
}
