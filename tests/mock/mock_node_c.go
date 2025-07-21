package mock

import (
	"context"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/core/node"
	"github.com/ljhe/scream/router"
)

type mockNodeC struct {
	*node.Node
}

func NewMockC(p iface.INodeBuilder) iface.INode {
	return &mockNodeC{
		Node: &node.Node{Id: p.GetID(), Ty: p.GetType(), Sys: p.GetSystem()},
	}
}

func (m *mockNodeC) Init(ctx context.Context) {
	m.Node.Init(ctx)

	m.OnEvent("ping", func(ctx iface.INodeContext) iface.IChain {
		return &node.DefaultChain{
			Handler: func(w *router.Wrapper) error {
				w.ToBuilder().WithResCustomFields(router.Attr{Key: "pong", Value: "pong"})
				return nil
			},
		}
	})

	m.OnEvent("test_block", func(ctx iface.INodeContext) iface.IChain {
		return &node.DefaultChain{
			Handler: func(w *router.Wrapper) error {

				val := router.GetReqCustomField[int](w, "randvalue")
				w.ToBuilder().WithResCustomFields(router.Attr{Key: "randvalue", Value: val + 1})

				return nil
			},
		}
	})
}
