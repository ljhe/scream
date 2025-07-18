package iface

import (
	"context"
	"sync"
)

type ISystem interface {
	Register(context.Context, INodeBuilder) (INode, error)
	Unregister(id, ty string) error

	// Loader returns the node loader
	Loader(string) INodeBuilder

	AddressBook() IAddressBook

	Exit(*sync.WaitGroup)
}
