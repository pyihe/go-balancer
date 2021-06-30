package balance

import (
	"sync"
)

type polling struct {
	sync.Mutex
	servers       nodeList
	existServers  map[string]struct{}
	currentWeight int64 // 当前所有权重之和
}

func NewPolling() *polling {
	return &polling{
		existServers: make(map[string]struct{}),
	}
}

func (p *polling) Next(ids ...string) Node {
	if len(p.servers) == 0 {
		return nil
	}

	p.Lock()
	defer p.Unlock()

	var nodeLen = len(p.servers)
	var result = p.servers[nodeLen-1]

	// 将权重最大的节点的权重减去当前总权重
	for i := range p.servers {
		var s = p.servers[i]
		var oldWeight = s.Weight()
		s.SetWeight(oldWeight * 2)
		if i == nodeLen-1 {
			s.SetWeight(s.Weight() - p.currentWeight)
		}
	}
	p.servers.Sort()
	p.currentWeight = 0
	for i := range p.servers {
		p.currentWeight += p.servers[i].Weight()
	}
	return result
}

func (p *polling) Remove(id string) (ok bool) {
	p.Lock()
	defer p.Unlock()
	for i := range p.servers {
		if p.servers[i].Id() == id {
			ok = true
			delete(p.existServers, id)
			p.currentWeight -= p.servers[i].Weight()
			p.servers = append(p.servers[:i], p.servers[i+1:]...)
			break
		}
	}
	return
}

func (p *polling) AddNode(node Node) {
	if node == nil || node.Weight() <= 0 {
		panic("nil node or negative weight")
	}
	p.Lock()
	defer p.Unlock()

	if _, ok := p.existServers[node.Id()]; ok {
		return
	}

	p.existServers[node.Id()] = struct{}{}
	p.servers = append(p.servers, node)
	p.currentWeight += node.Weight()
	p.servers.Sort()
}

func (p *polling) Update(id string, node Node) {
	if node == nil || node.Weight() <= 0 {
		panic("nil node or negative weight")
	}

	p.Lock()
	defer p.Unlock()

	var needSort bool
	for i := range p.servers {
		if p.servers[i].Id() == id {
			if p.servers[i].Weight() != node.Weight() {
				p.currentWeight += node.Weight() - p.servers[i].Weight()
				needSort = true
			}
			if node.Id() != id {
				delete(p.existServers, id)
				p.existServers[node.Id()] = struct{}{}
			}
			p.servers[i] = node
			break
		}
	}
	if needSort {
		p.servers.Sort()
	}
}

func (p *polling) Get(id string) (result Node) {
	for i := range p.servers {
		if p.servers[i].Id() == id {
			result = p.servers[i]
			break
		}
	}
	return
}