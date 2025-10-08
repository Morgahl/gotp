package process

import (
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/dbg"
	"github.com/Morgahl/gotp/internal/pid"
)

const (
	registry_DEFAULT_SIZE = 1024
)

var (
	localID uint64 = 0
	serial  uint32 = 0

	pidRegistryMu sync.RWMutex
	pidRegistry   map[PID]Ref
	removedPIDs   int

	nameRegistryMu sync.RWMutex
	nameRegistry   map[gotp.Atom]Ref
	removedNames   int

	gcInterval = 10 * time.Second
	lastGC     = time.Now().Add(time.Minute)
)

func init() {
	pidRegistry = make(map[PID]Ref, registry_DEFAULT_SIZE)
	nameRegistry = make(map[gotp.Atom]Ref, registry_DEFAULT_SIZE)
	go gcWaiter()
}

func nextPID() PID {
	id := atomic.AddUint64(&localID, 1)
	if id > pid.ID_MASK {
		stepSerial()
		id = 0
	}
	return pid.NewPID(0, id, uint8(atomic.LoadUint32(&serial)&pid.SERIAL_MASK))
}

func stepSerial() {
	atomic.AddUint32(&serial, 1)
	atomic.StoreUint64(&localID, 0)
}

func registerPID(pid PID, ref Ref) func() {
	pidRegistryMu.Lock()
	if _, exists := pidRegistry[pid]; exists {
		pidRegistryMu.Unlock()
		dbg.Throw("process with PID %s already registered", pid)
	}

	pidRegistry[pid] = ref
	pidRegistryMu.Unlock()
	return func() {
		pidRegistryMu.Lock()
		delete(pidRegistry, pid)
		removedPIDs++
		pidRegistryMu.Unlock()
	}
}

func sendPID(pid PID, msg Message) {
	pidRegistryMu.RLock()
	proc, exists := pidRegistry[pid]
	pidRegistryMu.RUnlock()
	if exists {
		proc.send(messageSignal(no_FLAGS, msg))
	}
}

func pidRef(pid PID) (Ref, bool) {
	pidRegistryMu.RLock()
	ref, exists := pidRegistry[pid]
	pidRegistryMu.RUnlock()
	if exists {
		return ref, true
	}
	return Ref{}, false
}

func registerNamed(name gotp.Atom, ref Ref) func() {
	nameRegistryMu.Lock()
	if _, exists := nameRegistry[name]; exists {
		nameRegistryMu.Unlock()
		dbg.Throw("process with name %s already registered", name)
	}
	nameRegistry[name] = ref
	nameRegistryMu.Unlock()
	return func() {
		nameRegistryMu.Lock()
		delete(nameRegistry, name)
		removedNames++
		nameRegistryMu.Unlock()
	}
}

func sendNamed(name gotp.Atom, msg Message) {
	nameRegistryMu.RLock()
	ref, exists := nameRegistry[name]
	nameRegistryMu.RUnlock()
	if exists {
		ref.send(messageSignal(no_FLAGS, msg))
	}
}

func namedRef(name gotp.Atom) (Ref, bool) {
	nameRegistryMu.RLock()
	ref, exists := nameRegistry[name]
	nameRegistryMu.RUnlock()
	if exists {
		return ref, true
	}
	return Ref{}, false
}

func gcWaiter() {
	for {
		if since := time.Since(lastGC); since < gcInterval {
			time.Sleep(gcInterval - since)
			continue
		}
		started := time.Now()
		pids, names := gc()
		slog.Warn("Garbage collection performed", slog.Int("pids", pids), slog.Int("names", names), slog.Duration("took", time.Since(started)))
	}
}

func gc() (pids, names int) {
	if time.Since(lastGC) < gcInterval {
		return
	}
	pids = gcPID()
	names = gcName()
	lastGC = time.Now()
	return
}

func gcPID() (count int) {
	pidRegistryMu.Lock()
	if removedPIDs == 0 || len(pidRegistry) <= registry_DEFAULT_SIZE {
		pidRegistryMu.Unlock()
		return count
	}
	newRegistry := make(map[PID]Ref, max(len(pidRegistry), registry_DEFAULT_SIZE))
	for k, v := range pidRegistry {
		newRegistry[k] = v
	}
	pidRegistry = newRegistry
	count = removedPIDs
	removedPIDs = 0
	pidRegistryMu.Unlock()
	return count
}

func gcName() (count int) {
	nameRegistryMu.Lock()
	if removedNames == 0 || len(nameRegistry) <= registry_DEFAULT_SIZE {
		nameRegistryMu.Unlock()
		return count
	}
	newNameRegistry := make(map[gotp.Atom]Ref, max(len(nameRegistry), registry_DEFAULT_SIZE))
	for k, v := range nameRegistry {
		newNameRegistry[k] = v
	}
	nameRegistry = newNameRegistry
	count = removedNames
	removedNames = 0
	nameRegistryMu.Unlock()
	return count
}
