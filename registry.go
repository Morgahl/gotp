package gotp

import (
	"log/slog"
	"sync"
)

const (
	REGISTRY_DEFAULT_SIZE = 100
)

type Registry[ID comparable] struct {
	mu    sync.RWMutex
	procs map[ID]Running
}

func New[ID comparable]() *Registry[ID] {
	return &Registry[ID]{
		procs: make(map[ID]Running, REGISTRY_DEFAULT_SIZE),
	}
}

func (r *Registry[ID]) Get(id ID) (p Running, exists bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, exists = r.procs[id]
	return
}

func (r *Registry[ID]) Put(id ID, p Running) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.procs[id] = p
}

func (r *Registry[ID]) Delete(id ID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	slog.Debug("Registry.Delete", "id", id)
	delete(r.procs, id)
}
