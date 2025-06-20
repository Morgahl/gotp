package ets

import (
	"sync"

	"github.com/Morgahl/gotp"
)

type Table[ID comparable, R any] struct {
	tname gotp.Atom
	tmu   sync.RWMutex
	tdata map[ID]R
}

func New[ID comparable, R any](name gotp.Atom) (*Table[ID, R], error) {
	t := Table[ID, R]{tname: name}
	if err := register(&t); err != nil {
		return nil, err
	}
	return &t, nil
}

func (t *Table[ID, R]) Name() gotp.Atom { return t.name() }
func (t *Table[ID, R]) name() gotp.Atom { return t.tname }

func (t *Table[ID, R]) GetAny(id ID) (r any, ok bool) { return t.get(id) }

func (t *Table[ID, R]) Get(id ID) (r R, ok bool) { return t.get(id) }
func (t *Table[ID, R]) get(id ID) (r R, ok bool) {
	t.tmu.RLock()
	if t.tdata == nil {
		t.tmu.RUnlock()
		return r, false
	}
	r, ok = t.tdata[id]
	t.tmu.RUnlock()
	return r, ok
}

func (t *Table[ID, R]) PutAny(id ID, r any) bool {
	if r, ok := r.(R); ok {
		t.put(id, r)
		return true
	}
	return false
}

func (t *Table[ID, R]) Put(id ID, r R) { t.put(id, r) }
func (t *Table[ID, R]) put(id ID, r R) {
	t.tmu.Lock()
	if t.tdata == nil {
		t.tdata = make(map[ID]R)
	}
	t.tdata[id] = r
	t.tmu.Unlock()
}
