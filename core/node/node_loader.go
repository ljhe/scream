package node

import (
	"context"
	"github.com/ljhe/scream/3rd/logrus"
	"github.com/ljhe/scream/core/iface"
)

type DefaultActorLoader struct {
	factory iface.INodeFactory
}

func BuildDefaultActorLoader(factory iface.INodeFactory) iface.INodeLoader {
	return &DefaultActorLoader{factory: factory}
}

// Builder selects an actor from the factory and provides a builder
func (dl *DefaultActorLoader) Builder(ty string, sys iface.ISystem) iface.INodeBuilder {
	ac := dl.factory.Get(ty)
	if ac == nil {
		return nil
	}

	builder := &NodeLoaderBuilder{
		ISystem:         sys,
		NodeConstructor: *ac,
		INodeLoader:     dl,
	}

	return builder
}

func (dl *DefaultActorLoader) AssignToNode(process iface.IProcess) {
	nodes := dl.factory.GetNodes()

	for _, node := range nodes {
		if node.Dynamic {
			continue
		}

		builder := dl.Builder(node.Name, process.System())
		if node.ID == "" {
			node.ID = node.Name
		}

		builder.WithID(process.ID() + "_" + node.ID)

		_, err := builder.Register(context.TODO())
		if err != nil {
			logrus.Errorf("assign to node build node %s err %v", node.Name, err)
		}
	}
}
