package node

import (
	"context"
	"errors"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/router"
)

type systemKey struct{}
type actorKey struct{}

type nodeContext struct {
	ctx context.Context
}

func (nc *nodeContext) Call(idOrSymbol, actorType, event string, mw *router.Wrapper) error {
	node, ok := nc.ctx.Value(actorKey{}).(iface.INode)
	if !ok {
		panic(errors.New("the node instance does not exist in the NodeContext"))
	}

	return node.Call(idOrSymbol, actorType, event, mw)
}

func (nc *nodeContext) Unregister(id, ty string) error {
	sys, ok := nc.ctx.Value(systemKey{}).(iface.ISystem)
	if !ok {
		panic(errors.New("the system instance does not exist in the NodeContext"))
	}

	return sys.Unregister(id, ty)
}
