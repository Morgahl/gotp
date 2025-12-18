package process

import (
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Morgahl/gotp"
)

const (
	DEFAULT_NAME_REGISTRY_SIZE = 1024
	PAGE_SIZE                  = 1024
)

// list manages pages of process references. Each page can hold up to 255 processes. When a page is full, a new page is
// created. Processes are assigned PIDs based on their position in the list. The list ensures that PIDs are unique and
// manages the allocation and deallocation of process references. When
type list struct {
	// this mutex is only held for page management related to spawning, not for page access of even gc as both are
	// eventually consistent operations downstream of process exit
	spawnMu sync.Mutex

	// total number of processes currently managed by the list
	processCount atomic.Uint64

	// [pid.raw % PAGE_SIZE] => page
	pages []*page

	// This mutex protects the name registry
	namesMu sync.RWMutex

	// map[gotp.Atom] => Ref
	names map[gotp.Atom]Ref
}

func newList() *list {
	ls := list{
		pages: make([]*page, 1, 16),
		names: make(map[gotp.Atom]Ref),
	}
	ls.pages[0] = newPage(1, ls.close(0), ls.handleExit)
	return &ls
}

func (ls *list) close(n int) func() {
	return func() {
		if n < 0 || n >= len(ls.pages) {
			return
		}
		ls.pages[n] = nil
	}
}

func (ls *list) handleExit() {
	ls.processCount.Add(^uint64(0))
}

func (ls *list) get(pid PID) Ref {
	page := pid.raw >> 8
	if int(page) >= len(ls.pages) {
		return Ref{}
	}
	pg := ls.pages[page]
	return pg.get(uint8(pid.raw & 0xFF))
}

func (ls *list) getNamed(name gotp.Atom) Ref {
	ls.namesMu.RLock()
	ref := ls.names[name]
	ls.namesMu.RUnlock()
	return ref
}

func (ls *list) spawn(fn RunFn, opts []SpawnOpt) Ref {
	if ls == nil {
		return Ref{}
	}

	ls.spawnMu.Lock()
	page := ls.pages[len(ls.pages)-1]

	ref, p, filled := page.build(opts)

	// if the page was filled during build, we need to create a new page
	if filled {
		nextIdx := len(ls.pages)
		ls.pages = append(ls.pages, newPage(uint64(nextIdx)*PAGE_SIZE+1, ls.close(nextIdx), ls.handleExit))
	}
	ls.spawnMu.Unlock()
	ls.processCount.Add(1)

	// register the name if needed
	if p.name != "" {
		ls.namesMu.Lock()
		ls.names[p.name] = ref
		ls.namesMu.Unlock()
		p.deregNameHandle = func() {
			ls.namesMu.Lock()
			delete(ls.names, p.name)
			ls.namesMu.Unlock()
		}
	}

	// start the process
	p.state = STARTED_STATE
	go p.run(fn)
	return ref
}

func (ls *list) status() {
	slog.Warn("PID Registry status", "count", ls.processCount.Load())

	pidStartTime := time.Now()
	clearedPages := 0
	for _, page := range ls.pages {
		if page == nil {
			clearedPages++
		}
	}
	took := time.Since(pidStartTime)
	slog.Warn("PID Page status", "took", took, "active", len(ls.pages)-clearedPages, "len", len(ls.pages))

	ls.namesMu.RLock()
	nameStartTime := time.Now()
	curLen := len(ls.names)
	ls.namesMu.RUnlock()
	took = time.Since(nameStartTime)
	slog.Warn("Name Registry status", "took", took, "count", curLen)
}

type page struct {
	// offset is the starting PID id for this page
	serial         uint8
	offset         uint64
	mu             sync.Mutex
	next, exited   uint64
	filled, closed bool
	close          func()
	handleExit     func()
	processes      [PAGE_SIZE]Ref
}

func newPage(offset uint64, close, handleExit func()) *page {
	pg := page{
		serial:     uint8(offset>>SERIAL_SHIFT) & SERIAL_MASK,
		offset:     offset >> ID_SHIFT & ID_MASK,
		close:      close,
		handleExit: handleExit,
	}
	return &pg
}

// get retrieves the process reference for the given module index within the page.
// WE EXPLICITLY DO NOT HOLD THE PAGE LOCK WHEN RETURNING THE REF, CALLER MUST HANDLE IT
func (pg *page) get(mod uint8) Ref {
	// check for nil and closed without lock as it only ever changes to true once
	if pg == nil || pg.closed {
		return Ref{}
	}
	// since we have an actual page at least lets return what we have at the index without lock
	return pg.processes[mod]
}

func (pg *page) build(opts []SpawnOpt) (ref Ref, p *process, filled bool) {
	pg.mu.Lock()
	idx := pg.next
	pg.next++
	// if next reaches PAGE_SIZE we mark the page as filled
	pg.filled = pg.next == PAGE_SIZE
	// unlock early to avoid holding lock during process building
	pg.mu.Unlock()

	// build the process
	p = build(NewPID(0, pg.offset+uint64(idx), pg.serial), opts)
	ref = newRef(p)
	p.deregPidHandle = func() {
		pg.processes[idx] = Ref{}
		pg.mu.Lock()
		pg.exited++
		pg.handleExit()
		// if exited reaches PAGE_SIZE we mark the page as closed
		pg.closed = pg.exited == PAGE_SIZE
		if pg.closed && pg.close != nil {
			pg.close()
			pg.close = nil
		}
		pg.mu.Unlock()
	}

	pg.processes[idx] = ref

	return ref, p, pg.filled
}
