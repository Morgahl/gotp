package process

import (
	"time"

	"github.com/Morgahl/gotp"
)

const (
	// 	registry_DEFAULT_SIZE = 1024
	status_INTERVAL = 10 * time.Second
)

var (
	// localID uint64 = 0
	// serial  uint32 = 0

	// pidRegistryMu sync.RWMutex
	// pidRegistry   map[PID]Ref
	// removedPIDs   int

	// nameRegistryMu sync.RWMutex
	// nameRegistry   map[gotp.Atom]Ref
	// removedNames   int

	registry *list

	lastStatus = time.Now().Add(time.Minute)
)

func init() {
	// pidRegistry = make(map[PID]Ref, registry_DEFAULT_SIZE)
	// nameRegistry = make(map[gotp.Atom]Ref, registry_DEFAULT_SIZE)
	registry = newList()
	go statusWaiter()
}

// func nextPID() PID {
// 	id := atomic.AddUint64(&localID, 1)
// 	if id > ID_MASK {
// 		stepSerial()
// 		id = 1 // skip 0 to never conflict with uninitialized PIDs
// 	}
// 	return NewPID(0, id, uint8(atomic.LoadUint32(&serial)&SERIAL_MASK))
// }

// func stepSerial() {
// 	atomic.AddUint32(&serial, 1)
// 	atomic.StoreUint64(&localID, 0)
// }

// func registerPID(ref Ref) func() {
// 	pidRegistryMu.Lock()
// 	if _, exists := pidRegistry[ref.pid]; exists {
// 		pidRegistryMu.Unlock()
// 		dbg.Throw("process with PID %s already registered", ref.pid)
// 	}

// 	pidRegistry[ref.pid] = ref
// 	pidRegistryMu.Unlock()
// 	return func() {
// 		pidRegistryMu.Lock()
// 		delete(pidRegistry, ref.pid)
// 		removedPIDs++
// 		pidRegistryMu.Unlock()
// 	}
// }

func sendPID(pid PID, msg gotp.Term) {
	// pidRegistryMu.RLock()
	// proc, exists := pidRegistry[pid]
	// pidRegistryMu.RUnlock()
	// if exists {
	// 	proc.send(messageSignal(no_FLAGS, msg))
	// }
	// registry.get(pid).send(messageSignal(no_FLAGS, msg))

	pidRef(pid).send(messageSignal(no_FLAGS, msg))
}

func pidRef(pid PID) Ref {
	// pidRegistryMu.RLock()
	// ref, exists := pidRegistry[pid]
	// pidRegistryMu.RUnlock()
	// if exists {
	// 	return ref, true
	// }
	// return Ref{}, false

	return registry.get(pid)
}

// func registerNamed(name gotp.Atom, ref Ref) func() {
// 	nameRegistryMu.Lock()
// 	if _, exists := nameRegistry[name]; exists {
// 		nameRegistryMu.Unlock()
// 		dbg.Throw("process with name %s already registered", name)
// 	}
// 	nameRegistry[name] = ref
// 	nameRegistryMu.Unlock()
// 	return func() {
// 		nameRegistryMu.Lock()
// 		delete(nameRegistry, name)
// 		removedNames++
// 		nameRegistryMu.Unlock()
// 	}
// }

func sendNamed(name gotp.Atom, msg gotp.Term) {
	// nameRegistryMu.RLock()
	// ref, exists := nameRegistry[name]
	// nameRegistryMu.RUnlock()
	//
	//	if exists {
	//		ref.send(messageSignal(no_FLAGS, msg))
	//	}
	namedRef(name).send(messageSignal(no_FLAGS, msg))
}

func namedRef(name gotp.Atom) Ref {
	// nameRegistryMu.RLock()
	// ref, exists := nameRegistry[name]
	// nameRegistryMu.RUnlock()
	// if exists {
	// 	return ref, true
	// }
	// return Ref{}, false

	return registry.getNamed(name)
}

func statusWaiter() {
	for {
		if since := time.Since(lastStatus); since < status_INTERVAL {
			time.Sleep(status_INTERVAL - since)
			continue
		}
		registry.status()
		lastStatus = time.Now()
	}
}

// func gc() (pids, names int) {
// 	if time.Since(lastGC) < gcInterval {
// 		return
// 	}
// 	pids = gcPID()
// 	names = gcName()
// 	lastGC = time.Now()
// 	return
// }

// func gcPID() (count int) {
// 	pidRegistryMu.Lock()
// 	if removedPIDs == 0 || len(pidRegistry) <= registry_DEFAULT_SIZE {
// 		pidRegistryMu.Unlock()
// 		return count
// 	}
// 	newRegistry := make(map[PID]Ref, max(len(pidRegistry), registry_DEFAULT_SIZE))
// 	for k, v := range pidRegistry {
// 		newRegistry[k] = v
// 	}
// 	pidRegistry = newRegistry
// 	count = removedPIDs
// 	removedPIDs = 0
// 	pidRegistryMu.Unlock()
// 	return count
// }

// func gcName() (count int) {
// 	nameRegistryMu.Lock()
// 	if removedNames == 0 || len(nameRegistry) <= registry_DEFAULT_SIZE {
// 		nameRegistryMu.Unlock()
// 		return count
// 	}
// 	newNameRegistry := make(map[gotp.Atom]Ref, max(len(nameRegistry), registry_DEFAULT_SIZE))
// 	for k, v := range nameRegistry {
// 		newNameRegistry[k] = v
// 	}
// 	nameRegistry = newNameRegistry
// 	count = removedNames
// 	removedNames = 0
// 	nameRegistryMu.Unlock()
// 	return count
// }
