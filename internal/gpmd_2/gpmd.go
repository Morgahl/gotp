package gpmd

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/internal/ctx"
)

type GPMD struct {
	ctx     ctx.Cancellable
	mu      sync.RWMutex
	nodeMap map[Node]time.Time
	hostMap map[string]Node
	nameMap map[gotp.Atom]Node
	appMap  map[string]Node
}

func New(ctx ctx.Cancellable) *GPMD {
	return &GPMD{
		ctx:     ctx,
		mu:      sync.RWMutex{},
		nodeMap: map[Node]time.Time{},
		hostMap: map[string]Node{},
		nameMap: map[gotp.Atom]Node{},
		appMap:  map[string]Node{},
	}
}

func (r *GPMD) Register(node Node, result *bool) error {
	slog.InfoContext(r.ctx, "Registering node", "node", node)
	r.mu.Lock()
	if _, ok := r.nodeMap[node]; ok {
		r.mu.Unlock()
		slog.ErrorContext(r.ctx, "Register: Node already registered", "node", node)
		*result = false
		return fmt.Errorf("node %s already registered", node)
	}

	r.nodeMap[node] = time.Now()
	r.nameMap[node.Name] = node
	r.hostMap[node.Host] = node
	r.mu.Unlock()
	slog.InfoContext(r.ctx, "Node registered", "node", node)
	*result = true

	return nil
}

func (r *GPMD) Unregister(node Node, result *bool) error {
	r.mu.Lock()
	if _, ok := r.nodeMap[node]; !ok {
		r.mu.Unlock()
		slog.ErrorContext(r.ctx, "Unregister: Node not registered", "node", node)
		*result = false
		return fmt.Errorf("node %s not registered", node)
	}

	delete(r.nodeMap, node)
	delete(r.nameMap, node.Name)
	delete(r.hostMap, node.Host)
	r.mu.Unlock()
	slog.InfoContext(r.ctx, "Node unregistered", "node", node)
	*result = true

	return nil
}

func (r *GPMD) Heartbeat(node Node, result *bool) error {
	r.mu.Lock()
	if _, ok := r.nodeMap[node]; !ok {
		r.mu.Unlock()
		slog.ErrorContext(r.ctx, "Heartbeat: Node not registered", "node", node)
		*result = false
		return fmt.Errorf("node %s not registered", node)
	}

	r.nodeMap[node] = time.Now()
	r.mu.Unlock()
	slog.InfoContext(r.ctx, "Heartbeat received for node", "node", node)
	*result = true

	return nil
}

func (r *GPMD) List(_ any, result *[]Node) error {
	r.mu.RLock()
	nodes := make([]Node, 0, len(r.nodeMap))
	for node := range r.nodeMap {
		nodes = append(nodes, node)
	}
	r.mu.RUnlock()
	slog.InfoContext(r.ctx, "Listing nodes", "count", len(nodes))
	*result = nodes
	return nil
}

func (r *GPMD) GetByName(name gotp.Atom, result *Node) error {
	r.mu.RLock()
	node, ok := r.nameMap[name]
	if !ok {
		r.mu.RUnlock()
		slog.ErrorContext(r.ctx, "GetByName: Node not found", "name", name)
		*result = Node{}
		return fmt.Errorf("node not found for name %s", name)
	}
	r.mu.RUnlock()
	slog.InfoContext(r.ctx, "Node found by name", "name", name, "node", node)
	*result = node
	return nil
}

func (r *GPMD) GetByApp(app string, result *Node) error {
	r.mu.RLock()
	node, ok := r.appMap[app]
	if !ok {
		r.mu.RUnlock()
		slog.ErrorContext(r.ctx, "GetByApp: Node not found", "app", app)
		*result = Node{}
		return fmt.Errorf("node not found for app %s", app)
	}
	r.mu.RUnlock()
	slog.InfoContext(r.ctx, "Node found by app", "app", app, "node", node)
	*result = node
	return nil
}

func (r *GPMD) GetByHost(host string, result *Node) error {
	r.mu.RLock()
	node, ok := r.hostMap[host]
	if !ok {
		r.mu.RUnlock()
		slog.ErrorContext(r.ctx, "GetByHost: Node not found", "host", host)
		*result = Node{}
		return fmt.Errorf("node not found for host %s", host)
	}
	r.mu.RUnlock()
	slog.InfoContext(r.ctx, "Node found by host", "host", host, "node", node)
	*result = node
	return nil
}
