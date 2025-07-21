package node

import (
	"context"
	"github.com/ljhe/scream/core/iface"
	"sync"
)

// NodeLoaderBuilder used to build NodeLoader
type NodeLoaderBuilder struct {
	iface.ISystem
	iface.NodeConstructor
	iface.INodeLoader

	optionsMutex sync.RWMutex
}

func (nlb *NodeLoaderBuilder) WithID(id string) iface.INodeBuilder {
	if id == "" {
		panic("actor id is empty")
	}
	nlb.ID = id
	return nlb
}

func (nlb *NodeLoaderBuilder) WithType(ty string) iface.INodeBuilder {
	nlb.Name = ty
	return nlb
}

func (nlb *NodeLoaderBuilder) WithOpt(key string, value string) iface.INodeBuilder {
	nlb.optionsMutex.Lock()
	nlb.Options[key] = value
	nlb.optionsMutex.Unlock()
	return nlb
}

func (nlb *NodeLoaderBuilder) GetID() string {
	return nlb.ID
}

func (nlb *NodeLoaderBuilder) GetType() string {
	return nlb.Name
}

func (nlb *NodeLoaderBuilder) GetWeight() int {
	return nlb.Weight
}

func (nlb *NodeLoaderBuilder) GetGlobalQuantityLimit() int {
	return nlb.GlobalQuantityLimit
}

func (nlb *NodeLoaderBuilder) GetNodeUnique() bool {
	return nlb.NodeUnique
}

func (nlb *NodeLoaderBuilder) GetSystem() iface.ISystem {
	return nlb.ISystem
}

func (nlb *NodeLoaderBuilder) GetLoader() iface.INodeLoader {
	return nlb.INodeLoader
}

func (nlb *NodeLoaderBuilder) GetConstructor() iface.CreateFunc {
	return nlb.Constructor
}

func (nlb *NodeLoaderBuilder) Register(ctx context.Context) (iface.INode, error) {
	return nlb.ISystem.Register(ctx, nlb)
}
