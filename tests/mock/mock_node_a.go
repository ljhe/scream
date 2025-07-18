package mock

import (
	"context"
	"fmt"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/core/node"
)

type mockActorA struct {
	*node.Node
}

func NewMockA(p iface.INodeBuilder) iface.INode {
	return &mockActorA{
		Node: &node.Node{Id: p.GetID(), Ty: p.GetType()},
	}
}

func (m *mockActorA) Init(ctx context.Context) {
	fmt.Println("mockActorA.Init")
}
