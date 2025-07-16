package agent

import (
	"github.com/Morgahl/gotp/process"
	"github.com/Morgahl/gotp/server"
)

func Start[T any](initFn InitFn[T], opts ...process.SpawnOpt) (*server.Server[any, any, T, any, any], error) {
	s, err := New(initFn).Start(opts...)
	if err != nil {
		return nil, err
	}
	return s.(*server.Server[any, any, T, any, any]), nil
}

func StartLink[T any](initFn InitFn[T], linked *process.Process, opts ...process.SpawnOpt) (*server.Server[any, any, T, any, any], error) {
	s, err := New(initFn).StartLink(linked, opts...)
	if err != nil {
		return nil, err
	}
	return s.(*server.Server[any, any, T, any, any]), nil
}

func Cast[T any, S process.Sendable](to S, msg UpdateFn[T]) {
	server.Cast[process.Message](to, msg)
}

func Get[T any, S process.Sendable](to S, from process.PID, msg GetFn[T]) (T, bool) {
	return server.Call[process.Message, T](to, from, msg, 0)
}

func GetAndUpdate[T any, S process.Sendable](to S, from process.PID, msg GetAndUpdateFn[T]) (T, bool) {
	return server.Call[process.Message, T](to, from, msg, 0)
}

func Update[T any, S process.Sendable](to S, msg UpdateFn[T]) {
	server.Cast[process.Message](to, msg)
}

func Stop[S process.Sendable](to S, reason error) {
	process.Send(to, server.StopMsg(reason))
}
