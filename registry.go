package gotp

import (
	"context"
	"sync"
)

const (
	REGISTRY_DEFAULT_SIZE = 100
)

type Registerable[ID comparable] interface {
	ID() ID
	Send(context.Context, Msg) error
}

type Registry[ID comparable] struct {
	mu    sync.RWMutex
	procs map[ID]Registerable[ID]
}

func New[ID comparable]() *Registry[ID] {
	return &Registry[ID]{
		procs: make(map[ID]Registerable[ID], REGISTRY_DEFAULT_SIZE),
	}
}

func (r *Registry[ID]) Get(id ID) (p Registerable[ID], exists bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, exists = r.procs[id]
	return
}

func (r *Registry[ID]) Put(p Registerable[ID]) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.procs[p.ID()] = p
}

func (r *Registry[ID]) Delete(p Registerable[ID]) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.procs, p.ID())
}
