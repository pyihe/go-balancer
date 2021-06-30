package balance

import "sort"

type BalancingType int8

const (
	Random BalancingType = iota + 1
	Polling
	Hash
)

type Balancer interface {
	AddNode(Node)
	Remove(id string) bool
	Update(id string, node Node)
	Next(ids ...string) Node
}

func NewBalancer(bType BalancingType) Balancer {
	switch bType {
	case Random:
		return NewRandom()
	case Polling:
		return NewPolling()
	default:
		return NewHash()
	}
}

type Node interface {
	Id() string
	Weight() int64
	SetWeight(int64)
}

type nodeList []Node

func (n nodeList) Len() int {
	return len(n)
}

func (n nodeList) Less(i, j int) bool {
	return n[i].Weight() < n[j].Weight()
}

func (n nodeList) Swap(i, j int) {
	n[i], n[j] = n[j], n[i]
}

func (n nodeList) Sort() {
	sort.Sort(n)
}
