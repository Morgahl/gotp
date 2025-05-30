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
	defer r.mu.RUnlock()
	p, exists = r.procs[id]
	return
}

func (r *Registry[ID]) Put(id ID, p Started) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.procs[id] = p
}

func (r *Registry[ID]) Delete(id ID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.procs, id)
}
