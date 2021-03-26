package balance

import (
	"sync"
)

type simplePolling struct {
	sync.RWMutex
	currentIndex int
	endpoints    []Node
}

func newSimplePolling() *simplePolling {
	return &simplePolling{}
}

func (p *simplePolling) GetNode(id string) interface{} {
	p.RLock()
	defer p.RUnlock()

	for _, n := range p.endpoints {
		if n.Id() == id {
			return n
		}
	}
	return nil
}

func (p *simplePolling) AddNode(node interface{}) {
	nt, ok := node.(Node)
	if !ok {
		return
	}
	p.Lock()
	defer p.Unlock()
	p.endpoints = append(p.endpoints, nt)
}

func (p *simplePolling) RemoveNode(nodeId string) {
	p.Lock()
	defer p.Unlock()

	for i, v := range p.endpoints {
		if nodeId == v.Id() {
			p.endpoints = append(p.endpoints[:i], p.endpoints[i+1:]...)
			break
		}
	}
}

func (p *simplePolling) Next(args ...interface{}) interface{} {
	p.RLock()
	defer p.RUnlock()

	total := len(p.endpoints)

	if p.currentIndex >= total {
		p.currentIndex = 0
	}
	node := p.endpoints[p.currentIndex]
	p.currentIndex = (p.currentIndex + 1) % total
	return node
}

type pollingWithWeight struct {
	sync.Mutex
	totalWeight   int
	currentWeight int
	endpoints     []WeightNode
}

func newPollingWithWeight() *pollingWithWeight {
	return &pollingWithWeight{}
}

func (p *pollingWithWeight) weight() (max int, total int) {
	for _, n := range p.endpoints {
		total += n.Weight()
		if n.Weight() > max {
			max = n.Weight()
		}
	}
	return
}

func (p *pollingWithWeight) GetNode(id string) interface{} {
	p.Lock()
	defer p.Unlock()
	for _, n := range p.endpoints {
		if n.Id() == id {
			return n
		}
	}
	return nil
}

func (p *pollingWithWeight) AddNode(node interface{}) {
	nt, ok := node.(WeightNode)
	if !ok {
		return
	}

	p.Lock()
	defer p.Unlock()

	p.endpoints = append(p.endpoints, nt)
	p.currentWeight, p.totalWeight = p.weight()
}

func (p *pollingWithWeight) RemoveNode(nodeId string) {
	p.Lock()
	defer p.Unlock()

	for i, n := range p.endpoints {
		if n.Id() == nodeId {
			p.endpoints = append(p.endpoints[:i], p.endpoints[i+1:]...)
			break
		}
	}
	p.currentWeight, p.totalWeight = p.weight()
}

func (p *pollingWithWeight) Next(args ...interface{}) interface{} {
	p.Lock()
	defer p.Unlock()

	// 每次找当前权重最大的节点
	var currentNode WeightNode
	for _, n := range p.endpoints {
		if n.Weight() == p.currentWeight {
			currentNode = n
			break
		}
	}

	if currentNode == nil {
		return nil
	}

	// 将权重最大的节点的权重减去当前总权重
	newWeight := currentNode.Weight() - p.totalWeight
	currentNode.UpdateWeight(newWeight)

	// 将操作后的节点列表中的每个节点的权重与初始权重相加
	for _, n := range p.endpoints {
		weight := n.Weight() + n.OriginalWeight()
		n.UpdateWeight(weight)
	}

	p.currentWeight, p.totalWeight = p.weight()
	return currentNode
}
