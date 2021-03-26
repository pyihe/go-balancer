package balance

type Balancer interface {
	AddNode(interface{})
	RemoveNode(node interface{})
	Next(...interface{}) interface{}
}

// 普通节点，非权重，非hash
type Node interface {
	Id() string
}

// 采用加权算法的节点，节点需要实现这个节点，如果采用的是非加权算法，实现空接口即可
type WeightNode interface {
	Node
	OriginalWeight() int // 初始权重
	Weight() int         // 获取节点当前权重
	UpdateWeight(int)    // 更新节点权重
}

// 采用一致性hash算法的节点
type HashNode interface {
	Identifier() string // 可以唯一标示节点的字符串
}

type Bt uint8

const (
	SimpleRandom Bt = iota + 1
	RandomWithWeight
	SimplePolling
	PollingWithWeight
	ConsistentHash
)

func NewBalancer(bt Bt) Balancer {
	switch bt {
	case SimpleRandom:
		return newSimpleRandom()
	case RandomWithWeight:
		return newRandomWithWeight()
	case SimplePolling:
		return newSimplePolling()
	case PollingWithWeight:
		return newPollingWithWeight()
	case ConsistentHash:
		return newHashMap()
	default:
		return nil
	}
}
