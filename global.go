package gotp

import (
	"fmt"
	"sync/atomic"
	"time"
)

var (
	localID     uint64 = 0
	serial      uint32 = 0
	pidRegistry Registry[PID, *Process]
)

func init() {
	pidRegistry = *New[PID, *Process]()
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

func register(p *Process) func() {
	pidRegistry.Put(p)
	return func() {
		pidRegistry.Delete(p)
	}
}

func Send(pid PID, msg Msg) {
	if proc, exists := pidRegistry.GetByID(pid); exists {
		proc.Send(msg)
	}
}

func SendNamed(name Atom, msg Msg) {
	if proc, exists := pidRegistry.GetByName(name); exists {
		proc.Send(msg)
	}
}

func SendAfter(pid PID, msg Msg, after time.Duration) *time.Timer {
	if proc, exists := pidRegistry.GetByID(pid); exists {
		return proc.SendAfter(msg, after)
	}
	return nil
}

func SendNamedAfter(name Atom, msg Msg, after time.Duration) *time.Timer {
	if proc, exists := pidRegistry.GetByName(name); exists {
		return proc.SendAfter(msg, after)
	}
	return nil
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
