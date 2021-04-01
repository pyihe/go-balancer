package balance

import (
	"math/rand"
	"sort"
	"sync"
	"time"
)

type simpleRandom struct {
	sync.RWMutex
	endpoints []Node
}

func newSimpleRandom() *simpleRandom {
	return &simpleRandom{}
}

func (s *simpleRandom) GetNode(id string) interface{} {
	s.RLock()
	defer s.RUnlock()

	for _, n := range s.endpoints {
		if n.Id() == id {
			return n
		}
	}
	return nil
}

func (s *simpleRandom) AddNode(node interface{}) {
	nt, ok := node.(Node)
	if !ok {
		return
	}

	s.Lock()
	defer s.Unlock()

	s.endpoints = append(s.endpoints, nt)
}

func (s *simpleRandom) RemoveNode(nodeId string) {
	s.Lock()
	defer s.Unlock()

	for i, n := range s.endpoints {
		if nodeId == n.Id() {
			s.endpoints = append(s.endpoints[:i], s.endpoints[i+1:]...)
		}
	}
}

func (s *simpleRandom) Next(args ...interface{}) interface{} {
	s.RLock()
	defer s.RUnlock()
	total := len(s.endpoints)
	if total == 0 {
		return nil
	}
	rand.Seed(time.Now().UnixNano())
	return s.endpoints[rand.Intn(total)]
}

func (s *simpleRandom) Range(f func(node Node) bool) {
	if len(s.endpoints) == 0 {
		return
	}

	for _, ep := range s.endpoints {
		if !f(ep) {
			break
		}
	}
}

type randomWithWeight struct {
	sync.RWMutex
	totalWeight int
	endpoints   []WeightNode
}

func newRandomWithWeight() *randomWithWeight {
	return &randomWithWeight{}
}

func (r *randomWithWeight) Len() int {
	return len(r.endpoints)
}

func (r *randomWithWeight) Less(i, j int) bool {
	return r.endpoints[i].Weight() < r.endpoints[j].Weight()
}

func (r *randomWithWeight) Swap(i, j int) {
	r.endpoints[i], r.endpoints[j] = r.endpoints[j], r.endpoints[i]
}

func (r *randomWithWeight) GetNode(id string) interface{} {
	r.RLock()
	defer r.RUnlock()
	for _, n := range r.endpoints {
		if n.Id() == id {
			return n
		}
	}
	return nil
}

func (r *randomWithWeight) AddNode(node interface{}) {
	tn, ok := node.(WeightNode)
	if !ok {
		return
	}

	r.Lock()
	defer r.Unlock()

	r.totalWeight += tn.Weight()
	r.endpoints = append(r.endpoints, tn)
}

func (r *randomWithWeight) RemoveNode(nodeId string) {
	r.Lock()
	defer r.Unlock()

	for i, n := range r.endpoints {
		if nodeId == n.Id() {
			r.totalWeight -= n.Weight()
			r.endpoints = append(r.endpoints[:i], r.endpoints[i+1:]...)
			break
		}
	}
}

func (r *randomWithWeight) Next(args ...interface{}) interface{} {
	r.RLock()
	defer r.RUnlock()

	var currentNode WeightNode
	rand.Seed(time.Now().UnixNano())
	weight := rand.Intn(r.totalWeight)
	sort.Sort(r)
	for i := 0; i < len(r.endpoints); i++ {
		if weight < r.endpoints[i].Weight() {
			currentNode = r.endpoints[i]
			break
		}
	}
	// 如果权重分配不均匀，则采用随机
	if currentNode == nil {
		idx := rand.Intn(len(r.endpoints))
		currentNode = r.endpoints[idx]
	}
	return currentNode
}

func (r *randomWithWeight) Range(f func(node WeightNode) bool) {
	if len(r.endpoints) == 0 {
		return
	}

	for _, ep := range r.endpoints {
		if !f(ep) {
			break
		}
	}
}
