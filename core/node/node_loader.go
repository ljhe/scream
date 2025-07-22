package node

import (
	"context"
	"github.com/ljhe/scream/3rd/logrus"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/def"
	"github.com/ljhe/scream/msg"
)

type DefaultNodeLoader struct {
	factory iface.INodeFactory
}

func BuildDefaultNodeLoader(factory iface.INodeFactory) iface.INodeLoader {
	return &DefaultNodeLoader{factory: factory}
}

func (nl *DefaultNodeLoader) Pick(ctx context.Context, builder iface.INodeBuilder) error {

	msgbuild := msg.NewBuilder(context.TODO())

	for key, value := range builder.GetOptions() {
		msgbuild.WithReqCustomFields(msg.Attr{Key: key, Value: value})
	}

	msgbuild.WithReqCustomFields(def.NodeID(builder.GetID()))
	msgbuild.WithReqCustomFields(def.NodeTy(builder.GetType()))

	go func() {
		err := builder.GetSystem().Call(def.SymbolWildcard, "MockDynamicPicker", "MockDynamicPick",
			msgbuild.Build(),
		)
		if err != nil {
			logrus.Warnf("nodeLoader call dynamic picker err %v", err.Error())
		}
	}()

	return nil
}

// Builder selects an actor from the factory and provides a builder
func (nl *DefaultNodeLoader) Builder(ty string, sys iface.ISystem) iface.INodeBuilder {
	ac := nl.factory.Get(ty)
	if ac == nil {
		return nil
	}

	builder := &NodeLoaderBuilder{
		ISystem:         sys,
		NodeConstructor: *ac,
		INodeLoader:     nl,
	}

	return builder
}

func (nl *DefaultNodeLoader) AssignToNode(process iface.IProcess) {
	nodes := nl.factory.GetNodes()

	for _, node := range nodes {
		if node.Dynamic {
			continue
		}

		builder := nl.Builder(node.Name, process.System())
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
