package mock

import (
	"github.com/ljhe/scream/core/iface"
)

// NodeFactory is a factory for creating nodes
type NodeFactory struct {
	Constructors map[string]*iface.NodeConstructor
}

// BuildNodeFactory create new node factory
func BuildNodeFactory() *NodeFactory {
	factory := &NodeFactory{
		Constructors: make(map[string]*iface.NodeConstructor),
	}

	factory.Constructors["mocka"] = &iface.NodeConstructor{
		ID:          "mocka",
		Name:        "mocka",
		Weight:      100,
		Constructor: NewMockA,
		NodeUnique:  false,
		Dynamic:     true,
		Options:     make(map[string]string),
	}

	return factory
}

func (factory *NodeFactory) Get(actorType string) *iface.NodeConstructor {
	if _, ok := factory.Constructors[actorType]; ok {
		return factory.Constructors[actorType]
	}

	return nil
}

func (factory *NodeFactory) GetActors() []*iface.NodeConstructor {
	var actors []*iface.NodeConstructor
	for _, v := range factory.Constructors {
		actors = append(actors, v)
	}
	return actors
}
