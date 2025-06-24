package gotp

import (
	"sync"
)

const (
	REGISTRY_DEFAULT_SIZE = 100
)

type Identifiable[V any] interface {
	ID() V
}

type Nameable[V any] interface {
	Name() Atom
}

type elem[V any] struct {
	name Atom
	v    V
}

type Registry[ID comparable, V Identifiable[ID]] struct {
	mu    sync.RWMutex
	procs map[ID]elem[V]
	names map[Atom]ID
}

func New[ID comparable, V Identifiable[ID]]() *Registry[ID, V] {
	return &Registry[ID, V]{
		procs: make(map[ID]elem[V], REGISTRY_DEFAULT_SIZE),
		names: make(map[Atom]ID, REGISTRY_DEFAULT_SIZE),
	}
}

func (r *Registry[ID, V]) GetByID(id ID) (p V, exists bool) {
	r.mu.RLock()
	elem, exists := r.procs[id]
	p = elem.v
	r.mu.RUnlock()
	return
}

func (r *Registry[ID, V]) GetByName(name Atom) (v V, exists bool) {
	var id ID
	var e elem[V]
	r.mu.RLock()
	id, exists = r.names[name]
	if exists {
		e, exists = r.procs[id]
	}
	r.mu.RUnlock()
	return e.v, exists
}

func (r *Registry[ID, V]) Put(v V) {
	id := v.ID()
	var name Atom
	if f, ok := any(v).(Nameable[V]); ok {
		name = f.Name()
	}
	r.mu.Lock()
	r.procs[id] = elem[V]{name: name, v: v}
	if name != "" {
		r.names[name] = id
	}
	r.mu.Unlock()
}

func (r *Registry[ID, V]) Delete(v V) {
	id := v.ID()
	r.mu.Lock()
	delete(r.procs, id)
	if f, ok := any(v).(Nameable[V]); ok {
		name := f.Name()
		if name != "" {
			delete(r.names, name)
		}
	}
	r.mu.Unlock()
}
