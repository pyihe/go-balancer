package balance

import (
	"crypto/sha1"
	"sync"

	"math"
	"sort"
	"strconv"
)

const (
	defaultVirtualSpots = 400
)

type nodesArray []node

func (p nodesArray) Len() int           { return len(p) }
func (p nodesArray) Less(i, j int) bool { return p[i].spotValue < p[j].spotValue }
func (p nodesArray) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p nodesArray) Sort()              { sort.Sort(p) }

type node struct {
	nodeKey   string
	spotValue uint32
}

type hash struct {
	mu           sync.RWMutex
	virtualSpots int
	nodes        nodesArray
	weights      map[string]Node
}

func NewHash() *hash {
	h := &hash{
		virtualSpots: defaultVirtualSpots,
		weights:      make(map[string]Node),
	}
	return h
}

func (h *hash) AddNode(node Node) {
	if node == nil || node.Weight() <= 0 {
		panic("nil node or negative weight")
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.weights[node.Id()] = node
	h.generate()
}

func (h *hash) Remove(id string) (ok bool) {
	if len(id) <= 0 {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.weights, id)
	h.generate()
	return true
}

func (h *hash) Update(id string, node Node) {
	if node == nil || node.Weight() <= 0 {
		panic("nil node or negative weight")
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.weights, id)
	h.weights[node.Id()] = node
	h.generate()
}

func (h *hash) Next(key ...string) Node {
	if len(key) == 0 {
		return nil
	}
	if len(h.nodes) == 0 {
		return nil
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	s := key[0]
	hasher := sha1.New()
	hasher.Write([]byte(s))
	hashBytes := hasher.Sum(nil)
	v := genValue(hashBytes[6:10])
	i := sort.Search(len(h.nodes), func(i int) bool { return h.nodes[i].spotValue >= v })

	if i == len(h.nodes) {
		i = 0
	}

	for k := range h.weights {
		if k == h.nodes[i].nodeKey {
			return h.weights[k]
		}
	}
	return nil
}

func (h *hash) generate() {
	h.nodes = nodesArray{}
	var totalW int64
	for _, e := range h.weights {
		totalW += e.Weight()
	}

	totalVirtualSpots := h.virtualSpots * len(h.weights)

	for id, n := range h.weights {
		spots := int(math.Floor(float64(n.Weight()) / float64(totalW) * float64(totalVirtualSpots)))
		for i := 1; i <= spots; i++ {
			hasher := sha1.New()
			hasher.Write([]byte(id + ":" + strconv.Itoa(i)))
			hashBytes := hasher.Sum(nil)
			newN := node{
				nodeKey:   id,
				spotValue: genValue(hashBytes[6:10]),
			}
			h.nodes = append(h.nodes, newN)
			hasher.Reset()
		}
	}
	h.nodes.Sort()
}

func genValue(bs []byte) uint32 {
	if len(bs) < 4 {
		return 0
	}
	v := (uint32(bs[3]) << 24) | (uint32(bs[2]) << 16) | (uint32(bs[1]) << 8) | (uint32(bs[0]))
	return v
}
