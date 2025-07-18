package iface

import "context"

type CreateFunc func(INodeBuilder) INode

type INode interface {
	Init(ctx context.Context)

	ID() string
	Type() string

	Exit()
}

type INodeLoader interface {
	// Builder selects a node from the factory and provides a builder
	Builder(string, ISystem) INodeBuilder
}

type INodeBuilder interface {
	GetID() string
	GetType() string
	GetGlobalQuantityLimit() int
	GetNodeUnique() bool
	GetWeight() int

	GetConstructor() CreateFunc

	WithID(string) INodeBuilder
	WithType(string) INodeBuilder
	WithOpt(string, string) INodeBuilder

	Register(context.Context) (INode, error)
}

type INodeContext interface {
	// Unregister unregisters an node
	Unregister(id, ty string) error
}

type NodeConstructor struct {
	ID   string
	Name string

	// Weight occupied by the actor, weight algorithm reference: 2c4g (pod = 2 * 4 * 1000)
	Weight int

	Dynamic bool

	// Constructor function
	Constructor CreateFunc

	// NodeUnique indicates whether this actor is unique within the current node
	NodeUnique bool

	// Global quantity limit for the current actor type that can be registered
	GlobalQuantityLimit int

	Options map[string]string
}

type INodeFactory interface {
	Get(ty string) *NodeConstructor
	GetActors() []*NodeConstructor
}
