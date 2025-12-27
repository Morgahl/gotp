package process

import (
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/term"
)

const (
	status_INTERVAL = 10 * time.Second
)

var (
	localID uint64 = 0
	serial  uint32 = 0

	pidTree  *PIDTree
	nameTree *NameTree

	lastStatus = time.Now().Add(status_INTERVAL)
)

func init() {
	pidTree = NewPIDTree()
	nameTree = NewNameTree()
	go statusWaiter()
}

func statusWaiter() {
	for {
		if since := time.Since(lastStatus); since < status_INTERVAL {
			time.Sleep(status_INTERVAL - since)
			continue
		}

		pids := pidTree.Count()
		names := nameTree.Count()

		slog.Warn("Registry Counts", "pids", pids, "names", names)

		lastStatus = time.Now()
	}
}

func nextPID() PID {
	id := atomic.AddUint64(&localID, 1)
	if id > ID_MASK {
		stepSerial()
	}
	return NewPID(0, id, uint8(atomic.LoadUint32(&serial)&SERIAL_MASK))
}

func stepSerial() {
	atomic.AddUint32(&serial, 1)
	atomic.StoreUint64(&localID, 1)
}

func sendPID(pid PID, msg term.Term) {
	pidRef(pid).send(messageSignal(no_FLAGS, msg))
}

func pidRef(pid PID) Ref {
	return pidTree.Load(pid.raw)
}

func sendNamed(name gotp.Atom, msg term.Term) {
	namedRef(name).send(messageSignal(no_FLAGS, msg))
}

func namedRef(name gotp.Atom) Ref {
	return nameTree.Load(name.String())
}
