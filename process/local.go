package process

import (
	"runtime"
	"sync"
	"sync/atomic"
	"time"

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

	lastGC     = time.Now()
	gcInterval = 60 * time.Second
)

func init() {
	pidRegistry = make(map[PID]Ref, registry_DEFAULT_SIZE)
	nameRegistry = make(map[gotp.Atom]Ref, registry_DEFAULT_SIZE)
	go gcWaiter()
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
			pidRegistryMu.Unlock()
			gcPID()
			return
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
		if removedNames*4 > len(nameRegistry) {
			nameRegistryMu.Unlock()
			gcName()
			return
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

func gcWaiter() {
	for {
		if since := time.Since(lastGC); since < gcInterval {
			time.Sleep(gcInterval - since)
			continue
		}
		gc()
	}
}

func gc() {
	if time.Since(lastGC) < gcInterval {
		return
	}

	gcPID()
	gcName()

	lastGC = time.Now()
	runtime.GC()
}

func gcPID() {
	pidRegistryMu.Lock()
	newRegistry := make(map[PID]Ref, max(len(pidRegistry), registry_DEFAULT_SIZE))
	for k, v := range pidRegistry {
		if v.IsValid() {
			newRegistry[k] = v
		}
	}
	pidRegistry = newRegistry
	removedPIDs = 0
	pidRegistryMu.Unlock()
}

func gcName() {
	nameRegistryMu.Lock()
	newNameRegistry := make(map[gotp.Atom]Ref, max(len(nameRegistry), registry_DEFAULT_SIZE))
	for k, v := range nameRegistry {
		if v.IsValid() {
			newNameRegistry[k] = v
		}
	}
	nameRegistry = newNameRegistry
	removedNames = 0
	nameRegistryMu.Unlock()
}
