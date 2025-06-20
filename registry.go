package gotp

import (
	"sync"
)

const (
	REGISTRY_DEFAULT_SIZE = 100
)

type Registry[ID comparable] struct {
	mu    sync.RWMutex
	procs map[ID]Started
}

func New[ID comparable]() *Registry[ID] {
	return &Registry[ID]{
		procs: make(map[ID]Started, REGISTRY_DEFAULT_SIZE),
	}
}

func (r *Registry[ID]) Get(id ID) (p Started, exists bool) {
	r.mu.RLock()
	p, exists = r.procs[id]
	r.mu.RUnlock()
	return
}

func (r *Registry[ID]) Put(id ID, p Started) {
	r.mu.Lock()
	r.procs[id] = p
	r.mu.Unlock()
}

func (r *Registry[ID]) Delete(id ID) {
	r.mu.Lock()
	delete(r.procs, id)
	r.mu.Unlock()
}
