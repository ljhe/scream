package socket

type NetServerNodeProperty struct {
	addr  string //
	name  string // 服务器名称
	zone  int    // 服务器区号
	typ   int    // 服务器类型
	index int    // 服务器区内的编号
}

func (n *NetServerNodeProperty) SetAddr(addr string) {
	n.addr = addr
}

func (n *NetServerNodeProperty) GetAddr() string {
	return n.addr
}

func (n *NetServerNodeProperty) SetName(s string) {
	n.name = s
}

func (n *NetServerNodeProperty) GetName() string {
	return n.name
}

func (n *NetServerNodeProperty) SetZone(z int) {
	n.zone = z
}

func (n *NetServerNodeProperty) GetZone() int {
	return n.zone
}

func (n *NetServerNodeProperty) SetServerTyp(t int) {
	n.typ = t
}

func (n *NetServerNodeProperty) GetServerTyp() int {
	return n.typ
}

func (n *NetServerNodeProperty) SetIndex(i int) {
	n.index = i
}

func (n *NetServerNodeProperty) GetIndex() int {
	return n.index
}
