package agent

import (
	gotp "github.com/Morgahl/gotp/old"
	"github.com/Morgahl/gotp/old/server"
)

func Start[T any](initFn InitFn[T], opts ...gotp.SpawnOpt) (*server.Server[any, any, T, any, any], error) {
	s, err := New(initFn).Start(opts...)
	if err != nil {
		return nil, err
	}
	return s.(*server.Server[any, any, T, any, any]), nil
}

func StartLink[T any](initFn InitFn[T], link gotp.PID, opts ...gotp.SpawnOpt) (*server.Server[any, any, T, any, any], error) {
	s, err := New(initFn).StartLink(link, opts...)
	if err != nil {
		return nil, err
	}
	return s.(*server.Server[any, any, T, any, any]), nil
}

func Cast[T any](to gotp.PID, msg UpdateFn[T]) {
	server.Cast[gotp.Msg](to, msg)
}

func Get[T any](to, from gotp.PID, msg GetFn[T]) (T, bool) {
	return server.Call[gotp.Msg, T](to, from, msg)
}

func GetAndUpdate[T any](to, from gotp.PID, msg GetAndUpdateFn[T]) (T, bool) {
	return server.Call[gotp.Msg, T](to, from, msg)
}

func Update[T any](to gotp.PID, msg UpdateFn[T]) {
	server.Cast[gotp.Msg](to, msg)
}
