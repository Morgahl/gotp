package gotp

import (
	"context"
	"log"
	"sync/atomic"
)

func init() {
	registry = *New[PID]()
}

var (
	pidCounter uint64 = 1
	registry   Registry[PID]
)

func nextPID() PID {
	return PID(atomic.AddUint64(&pidCounter, 1))
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
