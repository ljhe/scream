package plugins

import (
	"common/iface"
	"sync"
)

type MultiServerNode interface {
	AddNode(ed *ETCDServiceDesc, node iface.INetNode)
	GetNode(id string) iface.INetNode
	DelNode(id string)
}

// NetServerNode 服务发现的节点管理
type NetServerNode struct {
	nodeList map[string]iface.INetNode
	mu       sync.RWMutex
}

func NewMultiServerNode() *NetServerNode {
	return &NetServerNode{
		nodeList: make(map[string]iface.INetNode),
	}
}

func (n *NetServerNode) AddNode(ed *ETCDServiceDesc, node iface.INetNode) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	n.nodeList[ed.Id] = node
}

func (n *NetServerNode) GetNode(id string) iface.INetNode {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.nodeList[id]
}

func (n *NetServerNode) DelNode(id string) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	delete(n.nodeList, id)
}
