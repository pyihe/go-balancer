package balance

import (
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

type uint32Slice []uint32

func (p uint32Slice) Len() int           { return len(p) }
func (p uint32Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p uint32Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

const (
	defaultReplicas = 400
)

type Hash func([]byte) uint32

type hashMap struct {
	sync.Mutex
	hash     Hash
	replicas int
	keys     uint32Slice
	data     map[uint32]HashNode
}

func newHashMap() *hashMap {
	var m = &hashMap{
		Mutex:    sync.Mutex{},
		hash:     crc32.ChecksumIEEE,
		replicas: defaultReplicas,
		keys:     uint32Slice{},
		data:     make(map[uint32]HashNode),
	}
	return m
}

func (h *hashMap) AddNode(node interface{}) {
	h.Lock()
	defer h.Unlock()

	tn, ok := node.(HashNode)
	if ok {
		for i := 0; i < h.replicas; i++ {
			hv := h.hash([]byte(strconv.Itoa(i) + tn.Identifier()))
			h.keys = append(h.keys, hv)
			h.data[hv] = tn
		}
		sort.Sort(h.keys)
	}
}

func (h *hashMap) RemoveNode(key string) {
	h.Lock()
	defer h.Unlock()

	var hv = h.hash([]byte(key))
	var idx = sort.Search(len(h.keys), func(i int) bool {
		return h.keys[i] >= hv
	})

	delete(h.data, h.keys[uint32(idx%len(h.keys))])
}

func (h *hashMap) GetNode(key string) interface{} {
	h.Lock()
	defer h.Unlock()

	hv := h.hash([]byte(key))
	idx := sort.Search(len(h.keys), func(i int) bool {
		return h.keys[i] >= hv
	})
	return h.data[h.keys[uint32(idx%len(h.keys))]]
}

func (h *hashMap) Next(args ...interface{}) interface{} {
	var result []HashNode
	for _, arg := range args {
		str, ok := arg.(string)
		if !ok || len(str) == 0 {
			continue
		}
		hv := h.hash([]byte(str))
		idx := sort.Search(len(h.keys), func(i int) bool {
			return h.keys[i] >= hv
		})
		result = append(result, h.data[h.keys[uint32(idx%len(h.keys))]])
	}
	return result
}

func (h *hashMap) Range(f func(node HashNode) bool) {
	if len(h.data) == 0 {
		return
	}
	for _, ep := range h.data {
		if !f(ep) {
			break
		}
	}
}
