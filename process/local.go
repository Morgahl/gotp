package process

import (
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/debug"
)

const (
	registry_DEFAULT_SIZE = 100
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
)

func init() {
	pidRegistry = make(map[PID]Ref, registry_DEFAULT_SIZE)
	nameRegistry = make(map[gotp.Atom]Ref, registry_DEFAULT_SIZE)
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

func registerPID(pid PID, ref Ref) func() {
	pidRegistryMu.Lock()
	if _, exists := pidRegistry[pid]; exists {
		pidRegistryMu.Unlock()
		debug.Throw("process with PID %s already registered", pid)
	}

	pidRegistry[pid] = ref
	pidRegistryMu.Unlock()
	return func() {
		pidRegistryMu.Lock()
		delete(pidRegistry, pid)
		removedPIDs++
		if removedPIDs*10 > len(pidRegistry) {
			newRegistry := make(map[PID]Ref, max(len(pidRegistry), registry_DEFAULT_SIZE))
			for k, v := range pidRegistry {
				newRegistry[k] = v
			}
			pidRegistry = newRegistry
			removedPIDs = 0
			if len(pidRegistry) > 10000 {
				// optimistically call for a GC since this likely means a lot of processes have been removed
				// and otherwise the memory will not be reclaimed until the next GC cycle
				runtime.GC()
			}
		}
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
		debug.Throw("process with name %s already registered", name)
	}
	nameRegistry[name] = ref
	nameRegistryMu.Unlock()
	return func() {
		nameRegistryMu.Lock()
		delete(nameRegistry, name)
		removedNames++
		if removedNames*10 > len(nameRegistry) {
			newRegistry := make(map[gotp.Atom]Ref, max(len(nameRegistry), registry_DEFAULT_SIZE))
			for k, v := range nameRegistry {
				newRegistry[k] = v
			}
			nameRegistry = newRegistry
			removedNames = 0
			if len(nameRegistry) > 10000 {
				// optimistically call for a GC since this likely means a lot of processes have been removed
				// and otherwise the memory will not be reclaimed until the next GC cycle
				runtime.GC()
			}
		}
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
