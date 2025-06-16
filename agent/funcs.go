package agent

import (
	"time"

	"github.com/Morgahl/gotp"
)

func Start[Fn InitFn[T], T any](initFn Fn, timeout time.Duration, opts ...gotp.SpawnOpt) (gotp.Started, error) {
	return New(initFn).Start(timeout, opts...)
}

func StartLink[Fn InitFn[T], T any](initFn Fn, link gotp.PID, timeout time.Duration, opts ...gotp.SpawnOpt) (gotp.Supervised, error) {
	return New(initFn).StartLink(link, timeout, opts...)
}
