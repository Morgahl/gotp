package gpmd

import (
	"context"
	"fmt"
	"log"
	"net/netip"
	"sync"
	"time"

	"github.com/Morgahl/gotp/internal/ctx"
)

type GPMD struct {
	ctx       ctx.Cancellable
	addresses []netip.AddrPort
	mu        sync.RWMutex
	nodeMap   map[Node]time.Time
	hostMap   map[netip.AddrPort]Node
	nameMap   map[string]Node
}

func New(ctx ctx.Cancellable, addresses []netip.AddrPort) (*GPMD, error) {
	gpmd := &GPMD{
		ctx:       ctx,
		addresses: addresses,
		mu:        sync.RWMutex{},
		nodeMap:   map[Node]time.Time{},
		hostMap:   map[netip.AddrPort]Node{},
		nameMap:   map[string]Node{},
	}

	if err := gpmd.run(); err != nil {
		return nil, err
	}

	return gpmd, nil
}

func (r *GPMD) Register(node Node) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.checkCtxs(); err != nil {
		return err
	}

	if _, ok := r.nodeMap[node]; ok {
		return fmt.Errorf("node %s already registered", node)
	}

	r.nodeMap[node] = time.Now()
	r.nameMap[node.Name] = node
	r.hostMap[node.Host] = node

	return nil
}

func (r *GPMD) Unregister(node Node) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.checkCtxs(); err != nil {
		return err
	}

	if _, ok := r.nodeMap[node]; !ok {
		return fmt.Errorf("node %s not registered", node)
	}

	delete(r.nodeMap, node)
	delete(r.nameMap, node.Name)
	delete(r.hostMap, node.Host)

	return nil
}

func (r *GPMD) Heartbeat(node Node) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err := r.checkCtxs(); err != nil {
		return err
	}

	if _, ok := r.nodeMap[node]; !ok {
		return fmt.Errorf("node %s not registered", node)
	}

	r.nodeMap[node] = time.Now()

	return nil
}

func (r *GPMD) List() (NodeList, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.checkCtxs(); err != nil {
		return nil, err
	}

	nodes := make([]Node, 0, len(r.nodeMap))
	for node := range r.nodeMap {
		nodes = append(nodes, node)
	}

	return nodes, nil
}

func (r *GPMD) GetByName(name string) (Node, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.checkCtxs(); err != nil {
		return Node{}, err
	}

	node, ok := r.nameMap[name]
	if !ok {
		return Node{}, fmt.Errorf("node not found for name %s", name)
	}

	return node, nil
}

func (r *GPMD) GetByHost(host netip.AddrPort) (Node, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.checkCtxs(); err != nil {
		return Node{}, err
	}

	node, ok := r.hostMap[host]
	if !ok {
		return Node{}, fmt.Errorf("node not found for host %s", host)
	}

	return node, nil
}

func (r *GPMD) checkCtxs() error {
	select {
	case <-r.ctx.Done():
		return context.Cause(r.ctx)
	default:
		return nil
	}
}

func (r *GPMD) run() error {
	for _, addr := range r.addresses {
		log.Printf("listening on %s", addr)
		if _, err := NewServer(r.ctx, addr, r); err != nil {
			return err
		}
	}

	return nil
}
