package agent

import (
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/server"
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

func Cast[T any](to gotp.PID, msg UpdateFn[T], timeout time.Duration) error {
	return server.Cast[gotp.Msg](to, msg, timeout)
}

func Get[T any](to, from gotp.PID, msg GetFn[T], timeout time.Duration) (T, bool, error) {
	return server.Call[gotp.Msg, T](to, from, msg, timeout)
}

func GetAndUpdate[T any](to, from gotp.PID, msg GetAndUpdateFn[T], timeout time.Duration) (T, bool, error) {
	return server.Call[gotp.Msg, T](to, from, msg, timeout)
}

func Update[T any](to gotp.PID, msg UpdateFn[T], timeout time.Duration) error {
	return server.Cast[gotp.Msg](to, msg, timeout)
}
