package ets

import (
	"fmt"
	"sync"

	"github.com/Morgahl/gotp"
)

var (
	tablesMu = &sync.RWMutex{}
	tables   = make(map[gotp.Atom]any)
)

func register[ID comparable, R any](t *Table[ID, R]) (err error) {
	tablesMu.Lock()
	_, ok := tables[t.name()]
	if ok {
		tablesMu.Unlock()
		return fmt.Errorf("table %s already registered", t.name())
	}
	tables[t.name()] = t
	tablesMu.Unlock()
	return nil
}

func retrieveAny[ID comparable](name gotp.Atom) (t any, ok bool) {
	tablesMu.RLock()
	t, ok = tables[name]
	tablesMu.RUnlock()
	return t, ok
}

func retrieve[ID comparable, R any](name gotp.Atom) (t *Table[ID, R], ok bool) {
	tablesMu.RLock()
	v, found := tables[name]
	tablesMu.RUnlock()
	if !found {
		return nil, false
	}
	t, ok = v.(*Table[ID, R])
	return t, ok
}

func deregister(name gotp.Atom) {
	tablesMu.Lock()
	delete(tables, name)
	tablesMu.Unlock()
}
