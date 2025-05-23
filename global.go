package gotp

import (
	"context"
	"log"
	"sync/atomic"
)

var (
	localID  uint64 = 0
	serial   uint32 = 0
	registry Registry[PID]
	root     PID
)

func init() {
	registry = *New[PID]()
	root = nextPID()
}

func RootPID() PID {
	return root
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

func Send(ctx context.Context, pid PID, msg Msg) {
	if proc, exists := registry.Get(pid); exists {
		log.Printf("Send: %v -> %v", msg, pid)
		proc.Send(ctx, msg)
	}
}
