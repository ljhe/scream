package node

import (
	"context"
	"errors"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/msg"
)

type systemKey struct{}
type nodeKey struct{}

type nodeContext struct {
	ctx context.Context
}

func (nc *nodeContext) Call(idOrSymbol, nodeType, event string, mw *msg.Wrapper) error {
	node, ok := nc.ctx.Value(nodeKey{}).(iface.INode)
	if !ok {
		panic(errors.New("the node instance does not exist in the NodeContext"))
	}

	return node.Call(idOrSymbol, nodeType, event, mw)
}

func (nc *nodeContext) ReenterCall(idOrSymbol, nodeType, event string, mw *msg.Wrapper) iface.IFuture {
	node, ok := nc.ctx.Value(nodeKey{}).(iface.INode)
	if !ok {
		panic(errors.New("the node instance does not exist in the nodeContext"))
	}

	return node.ReenterCall(idOrSymbol, nodeType, event, mw)
}

func (nc *nodeContext) AddressBook() iface.IAddressBook {
	sys, ok := nc.ctx.Value(systemKey{}).(iface.ISystem)
	if !ok {
		panic(errors.New("the system instance does not exist in the NodeContext"))
	}

	return sys.AddressBook()
}

func (nc *nodeContext) Unregister(id, ty string) error {
	sys, ok := nc.ctx.Value(systemKey{}).(iface.ISystem)
	if !ok {
		panic(errors.New("the system instance does not exist in the NodeContext"))
	}

	return sys.Unregister(id, ty)
}
