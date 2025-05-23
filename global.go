package gotp

import (
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"
)

var (
	localID  uint64 = 0
	serial   uint32 = 0
	registry Registry[PID]
)

func init() {
	registry = *New[PID]()
}

func nextPID() PID {
	id := atomic.AddUint64(&localID, 1)
	if id > ID_MASK {
		stepSerial()
		id = 1
	}
	return newPID(0, id, uint8(atomic.LoadUint32(&serial)&SERIAL_MASK))
}

func stepSerial() {
	atomic.AddUint32(&serial, 1)
	atomic.StoreUint64(&localID, 0)
}

func register(p Registerable[PID]) func() {
	registry.Put(p)
	return func() {
		registry.Delete(p)
	}
}

func Send(pid PID, msg Msg, timeout time.Duration) error {
	if proc, exists := registry.Get(pid); exists {
		slog.Debug("global.Send", slog.String("pid", pid.String()), slog.String("msg", fmt.Sprintf("%v", msg)))
		return proc.Send(msg, timeout)
	}
	return NewUnknownPID(pid)
}

type UnknownPID struct {
	pid PID
}

func NewUnknownPID(pid PID) *UnknownPID {
	return &UnknownPID{pid: pid}
}

func (e *UnknownPID) Error() string {
	return fmt.Sprintf("UnknownPID{%v}", e.pid)
}

func (e *UnknownPID) Is(target error) bool {
	return target == e
}
