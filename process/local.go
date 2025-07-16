package process

import (
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
	pidRegistry   map[PID]*Process

	nameRegistryMu sync.RWMutex
	nameRegistry   map[gotp.Atom]Ref
)

func init() {
	pidRegistry = make(map[PID]*Process, registry_DEFAULT_SIZE)
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

func registerPID(p *Process) func() {
	pid := p.PID()
	pidRegistryMu.Lock()
	if _, exists := pidRegistry[pid]; exists {
		pidRegistryMu.Unlock()
		debug.Throw("Process with PID %s already registered", pid)
	}

	pidRegistry[pid] = p
	pidRegistryMu.Unlock()
	return func() {
		pidRegistryMu.Lock()
		delete(pidRegistry, pid)
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

func registerNamed(name gotp.Atom, ref Ref) func() {
	nameRegistryMu.Lock()
	if _, exists := nameRegistry[name]; exists {
		nameRegistryMu.Unlock()
		debug.Throw("Process with name %s already registered", name)
	}
	nameRegistry[name] = ref
	nameRegistryMu.Unlock()
	return func() {
		nameRegistryMu.Lock()
		delete(nameRegistry, name)
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

func namedPID(name gotp.Atom) (PID, bool) {
	nameRegistryMu.RLock()
	ref, exists := nameRegistry[name]
	nameRegistryMu.RUnlock()
	if exists {
		return ref.pid, true
	}
	return PIDZero(), false
}
