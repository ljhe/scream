package plugins

type MultiServerNode interface {
}

type NetServerNode struct {
}

func NewMultiServerNode() *NetServerNode {
	return &NetServerNode{}
}
