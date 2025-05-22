package socket

import (
	"github.com/ljhe/scream/core/config"
)

type NodeProp struct {
	addr  string //
	name  string // 服务器名称
	zone  int    // 服务器区号
	typ   int    // 服务器类型
	index int    // 服务器区内的编号
}

func (n *NodeProp) SetAddr(addr string) {
	n.addr = addr
}

func (n *NodeProp) GetAddr() string {
	return n.addr
}

func (n *NodeProp) SetName(s string) {
	n.name = s
}

func (n *NodeProp) GetName() string {
	return n.name
}

func (n *NodeProp) SetZone(z int) {
	n.zone = z
}

func (n *NodeProp) GetZone() int {
	return n.zone
}

func (n *NodeProp) SetServerTyp(t int) {
	n.typ = t
}

func (n *NodeProp) GetServerTyp() int {
	return n.typ
}

func (n *NodeProp) SetIndex(i int) {
	n.index = i
}

func (n *NodeProp) GetIndex() int {
	return n.index
}

func (n *NodeProp) SetNodeProp() {
	n.SetServerTyp(config.SConf.Node.Typ)
	n.SetZone(config.SConf.Node.Zone)
	n.SetIndex(config.SConf.Node.Index)
}
