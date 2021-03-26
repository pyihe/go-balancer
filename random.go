package balance

import (
	"math/rand"
	"reflect"
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

func (s *simpleRandom) AddNode(node interface{}) {
	s.Lock()
	defer s.Unlock()
	s.endpoints = append(s.endpoints, node)
}

func (s *simpleRandom) RemoveNode(node interface{}) {
	s.Lock()
	defer s.Unlock()
	for i, n := range s.endpoints {
		if reflect.DeepEqual(node, n) {
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

func (r *randomWithWeight) AddNode(node interface{}) {
	r.Lock()
	defer r.Unlock()

	if tn, ok := node.(WeightNode); ok {
		r.totalWeight += tn.Weight()
		r.endpoints = append(r.endpoints, tn)
	}
}

func (r *randomWithWeight) RemoveNode(node interface{}) {
	r.Lock()
	defer r.Unlock()

	for i, n := range r.endpoints {
		if reflect.DeepEqual(n, node) {
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
