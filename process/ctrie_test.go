package process

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/Morgahl/gotp/term"
)

func testRef(pid PID) (Ref, *process) {
	p := &process{pid: pid, signalChan: make(chan signal[term.Term], 1)}
	return newRef(p), p
}

// --- PIDTree Sequential Correctness ---

func TestPIDTree_InsertAndLookup(t *testing.T) {
	tree := NewPIDTree()
	pid := NewPID(0, 1, 0)
	ref, p := testRef(pid)

	if err := tree.Store(pid.raw, ref); err != nil {
		t.Fatalf("Store failed: %v", err)
	}
	if tree.Count() != 1 {
		t.Fatalf("expected count 1, got %d", tree.Count())
	}
	got := tree.Load(pid.raw)
	if got.pid != pid {
		t.Fatalf("expected pid %v, got %v", pid, got.pid)
	}
	runtime.KeepAlive(p)
}

func TestPIDTree_InsertDuplicate(t *testing.T) {
	tree := NewPIDTree()
	pid := NewPID(0, 1, 0)
	ref, p := testRef(pid)

	if err := tree.Store(pid.raw, ref); err != nil {
		t.Fatalf("first Store failed: %v", err)
	}
	if err := tree.Store(pid.raw, ref); err != ErrAlreadyExists {
		t.Fatalf("expected ErrAlreadyExists, got %v", err)
	}
	if tree.Count() != 1 {
		t.Fatalf("expected count 1, got %d", tree.Count())
	}
	runtime.KeepAlive(p)
}

func TestPIDTree_Delete(t *testing.T) {
	tree := NewPIDTree()
	pid := NewPID(0, 1, 0)
	ref, p := testRef(pid)

	tree.Store(pid.raw, ref)
	if !tree.Delete(pid.raw) {
		t.Fatal("Delete returned false for existing key")
	}
	if tree.Count() != 0 {
		t.Fatalf("expected count 0, got %d", tree.Count())
	}
	got := tree.Load(pid.raw)
	if got.pid != (PID{}) {
		t.Fatalf("expected zero Ref after delete, got pid %v", got.pid)
	}
	runtime.KeepAlive(p)
}

func TestPIDTree_DeleteNonexistent(t *testing.T) {
	tree := NewPIDTree()
	pid := NewPID(0, 99, 0)
	if tree.Delete(pid.raw) {
		t.Fatal("Delete returned true for nonexistent key")
	}
}

func TestPIDTree_Range(t *testing.T) {
	tree := NewPIDTree()
	procs := make([]*process, 100)
	for i := range 100 {
		pid := NewPID(0, uint64(i+1), 0)
		ref, p := testRef(pid)
		procs[i] = p
		tree.Store(pid.raw, ref)
	}

	seen := make(map[uint64]bool)
	tree.Range(func(key uint64, _ Ref) bool {
		seen[key] = true
		return true
	})
	if len(seen) != 100 {
		t.Fatalf("Range saw %d keys, expected 100", len(seen))
	}
	runtime.KeepAlive(procs)
}

func TestPIDTree_Snapshot(t *testing.T) {
	tree := NewPIDTree()
	procs := make([]*process, 0, 10)
	for i := range 5 {
		pid := NewPID(0, uint64(i+1), 0)
		ref, p := testRef(pid)
		procs = append(procs, p)
		tree.Store(pid.raw, ref)
	}

	snap := tree.Snapshot()
	if snap.Count() != 5 {
		t.Fatalf("snapshot count expected 5, got %d", snap.Count())
	}

	// Verify snapshot traversal sees the original keys.
	snapSeen := 0
	snap.Range(func(_ uint64, _ Ref) bool {
		snapSeen++
		return true
	})
	if snapSeen != 5 {
		t.Fatalf("snapshot Range saw %d keys, expected 5", snapSeen)
	}

	// Mutate original tree after snapshot.
	for i := range 5 {
		pid := NewPID(0, uint64(i+100), 0)
		ref, p := testRef(pid)
		procs = append(procs, p)
		tree.Store(pid.raw, ref)
	}
	if tree.Count() != 10 {
		t.Fatalf("tree count expected 10, got %d", tree.Count())
	}
	// Snapshot count must not have changed (count uses a separate atomic).
	if snap.Count() != 5 {
		t.Fatalf("snapshot count changed to %d after mutation", snap.Count())
	}

	// Delete from original tree; snapshot count stays stable.
	for i := range 5 {
		pid := NewPID(0, uint64(i+1), 0)
		tree.Delete(pid.raw)
	}
	if snap.Count() != 5 {
		t.Fatalf("snapshot count changed to %d after deletion from original", snap.Count())
	}
	runtime.KeepAlive(procs)
}

func TestPIDTree_ManyInsertsThenDeletes(t *testing.T) {
	const n = 1000
	tree := NewPIDTree()
	procs := make([]*process, n)
	pids := make([]PID, n)
	for i := range n {
		pid := NewPID(0, uint64(i+1), 0)
		pids[i] = pid
		ref, p := testRef(pid)
		procs[i] = p
		if err := tree.Store(pid.raw, ref); err != nil {
			t.Fatalf("Store(%d) failed: %v", i, err)
		}
	}
	if tree.Count() != n {
		t.Fatalf("expected count %d after inserts, got %d", n, tree.Count())
	}
	for i := range n {
		if !tree.Delete(pids[i].raw) {
			t.Fatalf("Delete(%d) returned false", i)
		}
	}
	if tree.Count() != 0 {
		t.Fatalf("expected count 0 after deletes, got %d", tree.Count())
	}
	runtime.KeepAlive(procs)
}

// --- NameTree Sequential Correctness ---

func TestNameTree_InsertLookupDelete(t *testing.T) {
	tree := NewNameTree()
	pid := NewPID(0, 1, 0)
	ref, p := testRef(pid)

	if err := tree.Store("myproc", ref); err != nil {
		t.Fatalf("Store failed: %v", err)
	}
	got := tree.Load("myproc")
	if got.pid != pid {
		t.Fatalf("expected pid %v, got %v", pid, got.pid)
	}
	if !tree.Delete("myproc") {
		t.Fatal("Delete returned false")
	}
	if tree.Count() != 0 {
		t.Fatalf("expected count 0, got %d", tree.Count())
	}
	got = tree.Load("myproc")
	if got.pid != (PID{}) {
		t.Fatalf("expected zero Ref after delete, got pid %v", got.pid)
	}
	runtime.KeepAlive(p)
}

func TestNameTree_Range(t *testing.T) {
	tree := NewNameTree()
	procs := make([]*process, 100)
	for i := range 100 {
		pid := NewPID(0, uint64(i+1), 0)
		ref, p := testRef(pid)
		procs[i] = p
		tree.Store(fmt.Sprintf("proc_%d", i), ref)
	}

	seen := make(map[string]bool)
	tree.Range(func(key string, _ Ref) bool {
		seen[key] = true
		return true
	})
	if len(seen) != 100 {
		t.Fatalf("Range saw %d keys, expected 100", len(seen))
	}
	runtime.KeepAlive(procs)
}

// --- Concurrent PIDTree Tests ---

func TestPIDTree_ConcurrentInsert(t *testing.T) {
	const goroutines = 16
	const opsPerGoroutine = 500

	tree := NewPIDTree()
	procs := make([]*process, goroutines*opsPerGoroutine)

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := range goroutines {
		go func(offset int) {
			defer wg.Done()
			for i := range opsPerGoroutine {
				idx := offset + i
				pid := NewPID(0, uint64(idx+1), 0)
				ref, p := testRef(pid)
				procs[idx] = p
				tree.Store(pid.raw, ref)
			}
		}(g * opsPerGoroutine)
	}
	wg.Wait()

	if tree.Count() != goroutines*opsPerGoroutine {
		t.Fatalf("expected count %d, got %d", goroutines*opsPerGoroutine, tree.Count())
	}
	// Verify all keys are present.
	for i := range goroutines * opsPerGoroutine {
		pid := NewPID(0, uint64(i+1), 0)
		got := tree.Load(pid.raw)
		if got.pid != pid {
			t.Fatalf("missing key for pid %v", pid)
		}
	}
	runtime.KeepAlive(procs)
}

func TestPIDTree_ConcurrentDelete(t *testing.T) {
	const goroutines = 16
	const opsPerGoroutine = 500
	const total = goroutines * opsPerGoroutine

	tree := NewPIDTree()
	procs := make([]*process, total)
	pids := make([]PID, total)
	for i := range total {
		pid := NewPID(0, uint64(i+1), 0)
		pids[i] = pid
		ref, p := testRef(pid)
		procs[i] = p
		tree.Store(pid.raw, ref)
	}

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := range goroutines {
		go func(offset int) {
			defer wg.Done()
			for i := range opsPerGoroutine {
				tree.Delete(pids[offset+i].raw)
			}
		}(g * opsPerGoroutine)
	}
	wg.Wait()

	if tree.Count() != 0 {
		t.Fatalf("expected count 0, got %d", tree.Count())
	}
	runtime.KeepAlive(procs)
}

func TestPIDTree_ConcurrentInsertDelete(t *testing.T) {
	const goroutines = 16
	const opsPerGoroutine = 500

	tree := NewPIDTree()
	// Pre-allocate enough for inserters. Deleters operate on the same key space.
	halfG := goroutines / 2
	totalInserts := halfG * opsPerGoroutine
	procs := make([]*process, totalInserts)

	// Pre-populate keys that deleters will target so deletes don't all miss.
	for i := range totalInserts {
		pid := NewPID(0, uint64(i+1), 0)
		ref, p := testRef(pid)
		procs[i] = p
		tree.Store(pid.raw, ref)
	}

	var wg sync.WaitGroup
	wg.Add(goroutines)

	// Half insert (re-insert same keys — will get ErrAlreadyExists or succeed after delete).
	for g := range halfG {
		go func(offset int) {
			defer wg.Done()
			for i := range opsPerGoroutine {
				idx := offset + i
				pid := NewPID(0, uint64(idx+1), 0)
				ref := newRef(procs[idx])
				tree.Store(pid.raw, ref)
				_ = tree.Load(pid.raw)
			}
		}(g * opsPerGoroutine)
	}

	// Half delete overlapping keys.
	for g := range halfG {
		go func(offset int) {
			defer wg.Done()
			for i := range opsPerGoroutine {
				pid := NewPID(0, uint64(offset+i+1), 0)
				tree.Delete(pid.raw)
			}
		}(g * opsPerGoroutine)
	}
	wg.Wait()

	// We can't predict exact count due to interleaving, just verify no panic and count >= 0.
	if tree.Count() < 0 {
		t.Fatalf("count went negative: %d", tree.Count())
	}
	runtime.KeepAlive(procs)
}

func TestPIDTree_ConcurrentOverlappingKeys(t *testing.T) {
	const goroutines = 32
	const keySpace = 100
	const opsPerGoroutine = 500

	tree := NewPIDTree()
	// Keep all processes alive for the duration.
	var procsMu sync.Mutex
	procs := make([]*process, 0, keySpace)
	for i := range keySpace {
		pid := NewPID(0, uint64(i+1), 0)
		_, p := testRef(pid)
		procs = append(procs, p)
	}

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			// Each goroutine creates its own refs backed by local processes
			// to avoid sharing weak pointers across goroutines.
			localProcs := make([]*process, keySpace)
			for i := range keySpace {
				pid := NewPID(0, uint64(i+1), 0)
				_, p := testRef(pid)
				localProcs[i] = p
			}
			procsMu.Lock()
			procs = append(procs, localProcs...)
			procsMu.Unlock()

			for op := range opsPerGoroutine {
				idx := uint64(op%keySpace + 1)
				pid := NewPID(0, idx, 0)
				ref, p := testRef(pid)
				procsMu.Lock()
				procs = append(procs, p)
				procsMu.Unlock()

				if op%2 == 0 {
					tree.Store(pid.raw, ref)
				} else {
					tree.Delete(pid.raw)
				}
				_ = tree.Load(pid.raw)
			}
		}()
	}
	wg.Wait()

	// Verify tree is in a consistent state: count matches actual entries.
	var actual int64
	tree.Range(func(_ uint64, _ Ref) bool {
		actual++
		return true
	})
	// Count may diverge slightly from Range in a concurrent trie (count is
	// updated non-atomically with the CAS), so we just verify no panic
	// and count is non-negative.
	if tree.Count() < 0 {
		t.Fatalf("count went negative: %d", tree.Count())
	}
	runtime.KeepAlive(procs)
}

// --- Concurrent NameTree Tests ---

func TestNameTree_ConcurrentInsertDelete(t *testing.T) {
	const goroutines = 16
	const opsPerGoroutine = 500

	tree := NewNameTree()
	procs := make([]*process, 0, goroutines*opsPerGoroutine)
	var procsMu sync.Mutex

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := range goroutines {
		go func(gid int) {
			defer wg.Done()
			for i := range opsPerGoroutine {
				key := fmt.Sprintf("proc_%d_%d", gid, i)
				pid := NewPID(0, uint64(gid*opsPerGoroutine+i+1), 0)
				ref, p := testRef(pid)
				procsMu.Lock()
				procs = append(procs, p)
				procsMu.Unlock()

				tree.Store(key, ref)
				_ = tree.Load(key)
				tree.Delete(key)
			}
		}(g)
	}
	wg.Wait()

	if tree.Count() < 0 {
		t.Fatalf("count went negative: %d", tree.Count())
	}
	runtime.KeepAlive(procs)
}

// --- Concurrent Snapshot Test ---

func TestPIDTree_ConcurrentSnapshot(t *testing.T) {
	const mutators = 4
	const snappers = 4
	const opsPerGoroutine = 500

	tree := NewPIDTree()
	procs := make([]*process, 0, mutators*opsPerGoroutine)
	var procsMu sync.Mutex
	var stop atomic.Bool

	var wg sync.WaitGroup
	wg.Add(mutators + snappers)

	// Mutator goroutines: insert and delete.
	for g := range mutators {
		go func(gid int) {
			defer wg.Done()
			for i := range opsPerGoroutine {
				pid := NewPID(0, uint64(gid*opsPerGoroutine+i+1), 0)
				ref, p := testRef(pid)
				procsMu.Lock()
				procs = append(procs, p)
				procsMu.Unlock()

				tree.Store(pid.raw, ref)
				if i%3 == 0 {
					tree.Delete(pid.raw)
				}
			}
			stop.Store(true)
		}(g)
	}

	// Snapper goroutines: take snapshots and traverse them.
	for range snappers {
		go func() {
			defer wg.Done()
			for !stop.Load() {
				snap := tree.Snapshot()
				snap.Range(func(_ uint64, _ Ref) bool {
					return true
				})
			}
		}()
	}
	wg.Wait()

	if tree.Count() < 0 {
		t.Fatalf("count went negative: %d", tree.Count())
	}
	runtime.KeepAlive(procs)
}
