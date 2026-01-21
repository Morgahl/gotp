package process

import (
	"expvar"
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
	localID          uint64 = 0
	globalSendCount  uint64 = 0
	processSendCount uint64 = 0

	pidTree  *PIDTree
	nameTree *NameTree

	lastStatus = time.Now().Add(status_INTERVAL)
)

func init() {
	pidTree = NewPIDTree()
	nameTree = NewNameTree()
	go statusWaiter()
	expvar.Publish("gotp_process", expvar.Func(collectMetrics))
}

func nextPID() PID {
	id := atomic.AddUint64(&localID, 1)
	return NewPID(0, id, uint8((id&SERIAL_MASK)>>SERIAL_SHIFT))
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

type Metrics struct {
	PIDCount         uint64
	NameCount        uint64
	GlobalSendCount  uint64
	ProcessSendCount uint64
}

func collectMetrics() any {
	return Metrics{
		PIDCount:         uint64(pidTree.Count()),
		NameCount:        uint64(nameTree.Count()),
		GlobalSendCount:  atomic.LoadUint64(&globalSendCount),
		ProcessSendCount: atomic.LoadUint64(&processSendCount),
	}
}
