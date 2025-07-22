package mock

import (
	"context"
	"fmt"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/core/node"
	"github.com/ljhe/scream/def"
	"github.com/ljhe/scream/msg"
	"time"
)

var MockCTccValue = 22

type mockNodeC struct {
	*node.Node
	tcc *TCC
}

func NewMockC(p iface.INodeBuilder) iface.INode {
	return &mockNodeC{
		Node: &node.Node{Id: p.GetID(), Ty: p.GetType(), Sys: p.GetSystem()},
		tcc:  &TCC{stateMap: make(map[string]*tccState)},
	}
}

func (m *mockNodeC) Init(ctx context.Context) {
	m.Node.Init(ctx)

	m.OnEvent("ping", func(ctx iface.INodeContext) iface.IChain {
		return &node.DefaultChain{
			Handler: func(w *msg.Wrapper) error {
				w.ToBuilder().WithResCustomFields(msg.Attr{Key: "pong", Value: "pong"})
				return nil
			},
		}
	})

	m.OnEvent("test_block", func(ctx iface.INodeContext) iface.IChain {
		return &node.DefaultChain{
			Handler: func(w *msg.Wrapper) error {

				val := msg.GetReqCustomField[int](w, "randvalue")
				w.ToBuilder().WithResCustomFields(msg.Attr{Key: "randvalue", Value: val + 1})

				return nil
			},
		}
	})

	m.OnEvent("tcc_succ", func(ctx iface.INodeContext) iface.IChain {
		return &node.DefaultChain{
			Handler: func(w *msg.Wrapper) error {

				transID := msg.GetReqCustomField[string](w, def.KeyTranscationID)

				m.tcc.stateMap[transID] = &tccState{
					originValue:  MockCTccValue,
					currentValue: 222,
					status:       "try",
					createdAt:    time.Now(),
				}

				MockCTccValue = 222
				fmt.Println("succ mock c value", MockCTccValue)
				return nil
			},
		}
	})

	m.OnEvent("tcc_confirm", func(ctx iface.INodeContext) iface.IChain {
		return &node.DefaultChain{
			Handler: func(w *msg.Wrapper) error {
				transID := msg.GetReqCustomField[string](w, def.KeyTranscationID)

				if state, exists := m.tcc.stateMap[transID]; exists {
					state.status = "confirmed"
					delete(m.tcc.stateMap, transID)
					return nil
				}
				return fmt.Errorf("transaction %s not found", transID)
			},
		}
	})
}
