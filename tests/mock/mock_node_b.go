package mock

import (
	"context"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/core/node"
	"github.com/ljhe/scream/router"
)

type mockNodeB struct {
	*node.Node
}

func NewMockB(p iface.INodeBuilder) iface.INode {
	return &mockNodeB{
		Node: &node.Node{Id: p.GetID(), Ty: p.GetType(), Sys: p.GetSystem()},
	}
}

func (m *mockNodeB) Init(ctx context.Context) {
	m.Node.Init(ctx)

	m.OnEvent("test_block", func(ctx iface.INodeContext) iface.IChain {
		return &node.DefaultChain{
			Handler: func(w *router.Wrapper) error {

				val := router.GetReqCustomField[int](w, "randvalue")
				w.ToBuilder().WithReqCustomFields(router.Attr{Key: "randvalue", Value: val + 1})
				ctx.Call("mockc", "mockc", "test_block", w)

				return nil
			},
		}
	})
}
