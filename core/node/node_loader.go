package node

import "github.com/ljhe/scream/core/iface"

type DefaultActorLoader struct {
	factory iface.INodeFactory
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

func BuildDefaultActorLoader(factory iface.INodeFactory) iface.INodeLoader {
	return &DefaultActorLoader{factory: factory}
}
