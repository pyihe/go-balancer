package balance

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type random struct {
	sync.RWMutex
	src          *rand.Rand
	existServers map[string]struct{}
	servers      nodeList // 根据权重从小到大
}

func NewRandom() *random {
	return &random{
		src:          rand.New(rand.NewSource(time.Now().UnixNano())),
		existServers: make(map[string]struct{}),
		servers:      nil,
	}
}

func (r *random) Next(ids ...string) Node {
	if len(r.servers) == 0 {
		return nil
	}
	var nodeLen = len(r.servers)
	var maxWeight = r.servers[nodeLen-1].Weight()
	fmt.Println(maxWeight)
	var targetWeight = r.src.Int63n(maxWeight + 1)
	var result Node

	r.RLock()
	defer r.RUnlock()
	for i := range r.servers {
		var s = r.servers[i]
		if s.Weight() >= targetWeight {
			result = r.servers[i]
			break
		}
	}
	return result
}

func (r *random) Remove(id string) (ok bool) {
	r.Lock()
	defer r.Unlock()
	for i := range r.servers {
		if r.servers[i].Id() == id {
			ok = true
			delete(r.existServers, id)
			r.servers = append(r.servers[:i], r.servers[i+1:]...)
			break
		}
	}
	return
}

func (r *random) AddNode(node Node) {
	if node == nil || node.Weight() <= 0 {
		panic("nil node or negative weight")
	}
	r.Lock()
	defer r.Unlock()

	if _, ok := r.existServers[node.Id()]; ok {
		return
	}

	r.existServers[node.Id()] = struct{}{}
	r.servers = append(r.servers, node)
	r.servers.Sort()
}

func (r *random) Update(id string, node Node) {
	if node == nil || node.Weight() <= 0 {
		panic("nil node or negative weight")
	}

	r.Lock()
	defer r.Unlock()

	var needSort bool
	for i := range r.servers {
		if r.servers[i].Id() == id {
			if r.servers[i].Weight() != node.Weight() {
				needSort = true
			}
			if node.Id() != id {
				delete(r.existServers, r.servers[i].Id())
				r.existServers[node.Id()] = struct{}{}
			}
			r.servers[i] = node
			break
		}
	}
	if needSort {
		r.servers.Sort()
	}
}
