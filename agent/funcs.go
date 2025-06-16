package agent

import (
	"github.com/Morgahl/gotp"
)

func Start[Fn InitFn[T], T any](initFn Fn, opts ...gotp.SpawnOpt) (gotp.Started, error) {
	return New(initFn).Start(opts...)
}

func StartLink[Fn InitFn[T], T any](initFn Fn, link gotp.PID, opts ...gotp.SpawnOpt) (gotp.Supervised, error) {
	return New(initFn).StartLink(link, opts...)
}
