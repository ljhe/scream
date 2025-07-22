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

	factory.Constructors["MockDynamicPicker"] = &iface.NodeConstructor{
		ID:                  "MockDynamicPicker",
		Name:                "MockDynamicPicker",
		Weight:              100,
		Constructor:         NewDynamicPickerActor,
		NodeUnique:          true,
		GlobalQuantityLimit: 10,
		Options:             make(map[string]string),
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

	factory.Constructors["mockb"] = &iface.NodeConstructor{
		ID:          "mockb",
		Name:        "mockb",
		Weight:      100,
		Constructor: NewMockB,
		NodeUnique:  false,
		Dynamic:     true,
		Options:     make(map[string]string),
	}

	factory.Constructors["mockc"] = &iface.NodeConstructor{
		ID:          "mockc",
		Name:        "mockc",
		Weight:      100,
		Constructor: NewMockC,
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

func (factory *NodeFactory) GetNodes() []*iface.NodeConstructor {
	var actors []*iface.NodeConstructor
	for _, v := range factory.Constructors {
		actors = append(actors, v)
	}
	return actors
}
