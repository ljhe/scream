package socket

type NetServerNodeProperty struct {
	addr string
}

func (n *NetServerNodeProperty) SetAddr(addr string) {
	n.addr = addr
}

func (n *NetServerNodeProperty) GetAddr() string {
	return n.addr
}
