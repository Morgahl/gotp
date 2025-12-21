package process

import (
	"math/bits"
	"sync/atomic"

	"github.com/Morgahl/gotp"
)

const (
	hamt_BITS = 8
	hamt_MASK = (1 << hamt_BITS) - 1

	ErrAlreadyExists = gotp.Atom("already_exists")
)

type hamtKey interface {
	~uint64 | ~string
}

type generation struct{ _ uint }

type PIDTree struct {
	root  *iNode[uint64]
	count atomic.Int64
}

type NameTree struct {
	root  *iNode[string]
	count atomic.Int64
}

func NewPIDTree() *PIDTree {
	return &PIDTree{root: &iNode[uint64]{gen: &generation{}}}
}

func NewNameTree() *NameTree {
	return &NameTree{root: &iNode[string]{gen: &generation{}}}
}

func (t *PIDTree) Count() int64  { return t.count.Load() }
func (t *NameTree) Count() int64 { return t.count.Load() }

func (t *PIDTree) Snapshot() *PIDTree {
	main := t.root.main.Load()
	newRoot := &iNode[uint64]{gen: &generation{}}
	newRoot.main.Store(main)
	snap := &PIDTree{root: newRoot}
	snap.count.Store(t.count.Load())
	return snap
}

func (t *NameTree) Snapshot() *NameTree {
	main := t.root.main.Load()
	newRoot := &iNode[string]{gen: &generation{}}
	newRoot.main.Store(main)
	snap := &NameTree{root: newRoot}
	snap.count.Store(t.count.Load())
	return snap
}

type iNode[K hamtKey] struct {
	gen  *generation
	main atomic.Pointer[mainNode[K]]
}

type mainNode[K hamtKey] struct {
	kind     nodeKind
	bitmap   [4]uint64
	children []*iNode[K]
	leaf     *leafNode[K]
	coll     *collNode[K]
}

type nodeKind uint8

const (
	kind_EMPTY nodeKind = iota
	kind_BRANCH
	kind_LEAF
	kind_COLLISION
)

type leafNode[K hamtKey] struct {
	hash  uint64
	key   K
	value Ref
}

type collNode[K hamtKey] struct {
	hash uint64
	keys []K
	vals []Ref
}

func (t *PIDTree) Store(key uint64, val Ref) error {
	h := pidHash(key)
	if ctrieInsert(t.root, t.root.gen, h, 0, key, val) {
		t.count.Add(1)
		return nil
	}
	return ErrAlreadyExists
}

func (t *PIDTree) Load(key uint64) Ref {
	h := pidHash(key)
	return ctrieLookup(t.root, t.root.gen, h, 0, key)
}

func (t *PIDTree) Delete(key uint64) bool {
	h := pidHash(key)
	if ctrieRemove(t.root, t.root.gen, h, 0, key) {
		t.count.Add(-1)
		return true
	}
	return false
}

func (t *PIDTree) Range(f func(uint64, Ref) bool) {
	snap := t.Snapshot()
	ctrieTraverse(snap.root, f)
}

func (t *NameTree) Store(key string, val Ref) error {
	h := nameHash(key)
	if ctrieInsert(t.root, t.root.gen, h, 0, key, val) {
		t.count.Add(1)
		return nil
	}
	return ErrAlreadyExists
}

func (t *NameTree) Load(key string) Ref {
	h := nameHash(key)
	return ctrieLookup(t.root, t.root.gen, h, 0, key)
}

func (t *NameTree) Delete(key string) bool {
	h := nameHash(key)
	if ctrieRemove(t.root, t.root.gen, h, 0, key) {
		t.count.Add(-1)
		return true
	}
	return false
}

func (t *NameTree) Range(f func(string, Ref) bool) {
	snap := t.Snapshot()
	ctrieTraverse(snap.root, f)
}

func getIndex(bitmap [4]uint64, bitIdx uint) int {
	word := bitIdx >> 6
	mask := uint64(1) << (bitIdx & 63)
	count := 0
	for i := uint(0); i < word; i++ {
		count += bits.OnesCount64(bitmap[i])
	}
	count += bits.OnesCount64(bitmap[word] & (mask - 1))
	return count
}

func (m *mainNode[K]) compressed() *mainNode[K] {
	if m.kind != kind_BRANCH || len(m.children) != 1 {
		return m
	}
	childMain := m.children[0].main.Load()
	if childMain != nil && childMain.kind == kind_LEAF {
		return childMain
	}
	return m
}

func ctrieInsert[K hamtKey](i *iNode[K], gen *generation, h uint64, shift uint, key K, val Ref) bool {
	for {
		main := i.main.Load()
		if i.gen != gen {
			main = ctrieCopy(i, gen)
		}
		if main == nil || main.kind == kind_EMPTY {
			newMain := &mainNode[K]{kind: kind_LEAF, leaf: &leafNode[K]{h, key, val}}
			if i.main.CompareAndSwap(main, newMain) {
				return true
			}
			continue
		}
		switch main.kind {
		case kind_LEAF:
			if main.leaf.hash == h && main.leaf.key == key {
				return false
			}
			var newMain *mainNode[K]
			if main.leaf.hash == h || shift >= 64-hamt_BITS {
				newMain = &mainNode[K]{
					kind: kind_COLLISION,
					coll: &collNode[K]{hash: h, keys: []K{main.leaf.key, key}, vals: []Ref{main.leaf.value, val}},
				}
			} else {
				newMain = ctrieJoin(gen, main.leaf, &leafNode[K]{h, key, val}, shift)
			}
			if i.main.CompareAndSwap(main, newMain) {
				return true
			}
		case kind_COLLISION:
			if main.coll.hash == h || shift >= 64-hamt_BITS {
				for _, k := range main.coll.keys {
					if k == key {
						return false
					}
				}
				nk, nv := make([]K, len(main.coll.keys)+1), make([]Ref, len(main.coll.vals)+1)
				copy(nk, main.coll.keys)
				copy(nv, main.coll.vals)
				nk[len(main.coll.keys)], nv[len(main.coll.vals)] = key, val
				newMain := &mainNode[K]{kind: kind_COLLISION, coll: &collNode[K]{hash: h, keys: nk, vals: nv}}
				if i.main.CompareAndSwap(main, newMain) {
					return true
				}
				continue
			}
			newMain := ctrieJoinLeafAndColl(gen, main.coll, &leafNode[K]{h, key, val}, shift)
			if i.main.CompareAndSwap(main, newMain) {
				return true
			}
		case kind_BRANCH:
			bitIdx := uint((h >> shift) & hamt_MASK)
			word := bitIdx >> 6
			mask := uint64(1) << (bitIdx & 63)
			if main.bitmap[word]&mask == 0 {
				child := &iNode[K]{gen: gen}
				child.main.Store(&mainNode[K]{kind: kind_LEAF, leaf: &leafNode[K]{h, key, val}})
				idx := getIndex(main.bitmap, bitIdx)
				newChildren := make([]*iNode[K], len(main.children)+1)
				copy(newChildren[:idx], main.children[:idx])
				newChildren[idx] = child
				copy(newChildren[idx+1:], main.children[idx:])
				newBitmap := main.bitmap
				newBitmap[word] |= mask
				newMain := &mainNode[K]{kind: kind_BRANCH, bitmap: newBitmap, children: newChildren}
				if i.main.CompareAndSwap(main, newMain) {
					return true
				}
			} else {
				idx := getIndex(main.bitmap, bitIdx)
				return ctrieInsert(main.children[idx], gen, h, shift+hamt_BITS, key, val)
			}
		}
	}
}

func ctrieLookup[K hamtKey](i *iNode[K], gen *generation, h uint64, shift uint, key K) Ref {
	main := i.main.Load()
	if i.gen != gen {
		main = ctrieCopy(i, gen)
	}
	if main == nil {
		return Ref{}
	}
	switch main.kind {
	case kind_LEAF:
		if main.leaf.hash == h && main.leaf.key == key {
			return main.leaf.value
		}
	case kind_COLLISION:
		if main.coll.hash == h {
			for idx, k := range main.coll.keys {
				if k == key {
					return main.coll.vals[idx]
				}
			}
		}
	case kind_BRANCH:
		bitIdx := uint((h >> shift) & hamt_MASK)
		word := bitIdx >> 6
		mask := uint64(1) << (bitIdx & 63)
		if main.bitmap[word]&mask != 0 {
			idx := getIndex(main.bitmap, bitIdx)
			return ctrieLookup(main.children[idx], gen, h, shift+hamt_BITS, key)
		}
	}
	return Ref{}
}

func ctrieRemove[K hamtKey](i *iNode[K], gen *generation, h uint64, shift uint, key K) bool {
	for {
		main := i.main.Load()
		if i.gen != gen {
			main = ctrieCopy(i, gen)
		}
		if main == nil || main.kind == kind_EMPTY {
			return false
		}
		switch main.kind {
		case kind_LEAF:
			if main.leaf.hash == h && main.leaf.key == key {
				if i.main.CompareAndSwap(main, &mainNode[K]{kind: kind_EMPTY}) {
					return true
				}
				continue
			}
			return false
		case kind_COLLISION:
			if main.coll.hash != h {
				return false
			}
			idx := -1
			for idxKey, k := range main.coll.keys {
				if k == key {
					idx = idxKey
					break
				}
			}
			if idx == -1 {
				return false
			}
			var newMain *mainNode[K]
			if len(main.coll.keys) == 2 {
				keep := 1 - idx
				newMain = &mainNode[K]{kind: kind_LEAF, leaf: &leafNode[K]{h, main.coll.keys[keep], main.coll.vals[keep]}}
			} else {
				nk, nv := make([]K, len(main.coll.keys)-1), make([]Ref, len(main.coll.vals)-1)
				copy(nk, main.coll.keys[:idx])
				copy(nk[idx:], main.coll.keys[idx+1:])
				copy(nv, main.coll.vals[:idx])
				copy(nv[idx:], main.coll.vals[idx+1:])
				newMain = &mainNode[K]{kind: kind_COLLISION, coll: &collNode[K]{hash: h, keys: nk, vals: nv}}
			}
			if i.main.CompareAndSwap(main, newMain) {
				return true
			}
		case kind_BRANCH:
			bitIdx := uint((h >> shift) & hamt_MASK)
			word := bitIdx >> 6
			mask := uint64(1) << (bitIdx & 63)
			if main.bitmap[word]&mask == 0 {
				return false
			}
			childIdx := getIndex(main.bitmap, bitIdx)
			child := main.children[childIdx]
			if !ctrieRemove(child, gen, h, shift+hamt_BITS, key) {
				return false
			}
			childMain := child.main.Load()
			if childMain != nil && childMain.kind == kind_EMPTY {
				newBitmap := main.bitmap
				newBitmap[word] &= ^mask
				var newMain *mainNode[K]
				if (newBitmap[0] | newBitmap[1] | newBitmap[2] | newBitmap[3]) == 0 {
					newMain = &mainNode[K]{kind: kind_EMPTY}
				} else {
					newChildren := make([]*iNode[K], len(main.children)-1)
					copy(newChildren[:childIdx], main.children[:childIdx])
					copy(newChildren[childIdx:], main.children[childIdx+1:])
					newMain = (&mainNode[K]{kind: kind_BRANCH, bitmap: newBitmap, children: newChildren}).compressed()
				}
				i.main.CompareAndSwap(main, newMain)
			}
			return true
		}
	}
}

func ctrieTraverse[K hamtKey](i *iNode[K], f func(K, Ref) bool) bool {
	main := i.main.Load()
	if main == nil {
		return true
	}
	switch main.kind {
	case kind_LEAF:
		return f(main.leaf.key, main.leaf.value)
	case kind_COLLISION:
		for idx := range main.coll.keys {
			if !f(main.coll.keys[idx], main.coll.vals[idx]) {
				return false
			}
		}
	case kind_BRANCH:
		for _, child := range main.children {
			if !ctrieTraverse(child, f) {
				return false
			}
		}
	}
	return true
}

func ctrieCopy[K hamtKey](i *iNode[K], gen *generation) *mainNode[K] {
	main := i.main.Load()
	i.main.Store(main)
	i.gen = gen
	return main
}

func ctrieJoin[K hamtKey](gen *generation, oldLeaf, newLeaf *leafNode[K], shift uint) *mainNode[K] {
	os, ns := uint((oldLeaf.hash>>shift)&hamt_MASK), uint((newLeaf.hash>>shift)&hamt_MASK)
	if os != ns {
		inOld, inNew := &iNode[K]{gen: gen}, &iNode[K]{gen: gen}
		inOld.main.Store(&mainNode[K]{kind: kind_LEAF, leaf: oldLeaf})
		inNew.main.Store(&mainNode[K]{kind: kind_LEAF, leaf: newLeaf})
		children := make([]*iNode[K], 2)
		if os < ns {
			children[0], children[1] = inOld, inNew
		} else {
			children[0], children[1] = inNew, inOld
		}
		var bm [4]uint64
		bm[os>>6] |= (1 << (os & 63))
		bm[ns>>6] |= (1 << (ns & 63))
		return &mainNode[K]{kind: kind_BRANCH, bitmap: bm, children: children}
	}
	childINode := &iNode[K]{gen: gen}
	childINode.main.Store(ctrieJoin(gen, oldLeaf, newLeaf, shift+hamt_BITS))
	var bm [4]uint64
	bm[os>>6] |= (1 << (os & 63))
	return &mainNode[K]{kind: kind_BRANCH, bitmap: bm, children: []*iNode[K]{childINode}}
}

func ctrieJoinLeafAndColl[K hamtKey](gen *generation, coll *collNode[K], leaf *leafNode[K], shift uint) *mainNode[K] {
	os, ns := uint((coll.hash>>shift)&hamt_MASK), uint((leaf.hash>>shift)&hamt_MASK)
	if os != ns {
		inColl, inLeaf := &iNode[K]{gen: gen}, &iNode[K]{gen: gen}
		inColl.main.Store(&mainNode[K]{kind: kind_COLLISION, coll: coll})
		inLeaf.main.Store(&mainNode[K]{kind: kind_LEAF, leaf: leaf})
		children := make([]*iNode[K], 2)
		if os < ns {
			children[0], children[1] = inColl, inLeaf
		} else {
			children[0], children[1] = inLeaf, inColl
		}
		var bm [4]uint64
		bm[os>>6] |= (1 << (os & 63))
		bm[ns>>6] |= (1 << (ns & 63))
		return &mainNode[K]{kind: kind_BRANCH, bitmap: bm, children: children}
	}
	childINode := &iNode[K]{gen: gen}
	childINode.main.Store(ctrieJoinLeafAndColl(gen, coll, leaf, shift+hamt_BITS))
	var bm [4]uint64
	bm[os>>6] |= (1 << (os & 63))
	return &mainNode[K]{kind: kind_BRANCH, bitmap: bm, children: []*iNode[K]{childINode}}
}

func pidHash(x uint64) uint64 {
	x ^= 0xa0761d6478bd642f
	low, high := bits.Mul64(x, 0xe7037ed1a0b428db)
	return low ^ high
}

func nameHash(s string) uint64 {
	h := uint64(0xcbf29ce484222325)
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= 0x100000001b3
	}
	return pidHash(h)
}
