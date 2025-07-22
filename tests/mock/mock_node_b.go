package mock

import (
	"context"
	"fmt"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/core/node"
	"github.com/ljhe/scream/def"
	"github.com/ljhe/scream/msg"
	"sync/atomic"
	"time"
)

var MockBTccValue = 11
var BechmarkCallReceivedMessageCount int64

type mockNodeB struct {
	*node.Node
	tcc *TCC
}

func NewMockB(p iface.INodeBuilder) iface.INode {
	return &mockNodeB{
		Node: &node.Node{Id: p.GetID(), Ty: p.GetType(), Sys: p.GetSystem()},
		tcc:  &TCC{stateMap: make(map[string]*tccState)},
	}
}

func (m *mockNodeB) Init(ctx context.Context) {
	m.Node.Init(ctx)

	m.OnEvent("clac", func(ctx iface.INodeContext) iface.IChain {
		return &node.DefaultChain{
			Handler: func(w *msg.Wrapper) error {

				val := msg.GetReqCustomField[int](w, "calculateVal")
				w.ToBuilder().WithResCustomFields(msg.Attr{Key: "calculateVal", Value: val + 2})

				return nil
			},
		}
	})

	m.OnEvent("call_benchmark", func(ctx iface.INodeContext) iface.IChain {
		return &node.DefaultChain{
			Handler: func(w *msg.Wrapper) error {
				atomic.AddInt64(&BechmarkCallReceivedMessageCount, 1)
				return nil
			},
		}
	})

	m.OnEvent("timeout", func(ctx iface.INodeContext) iface.IChain {
		return &node.DefaultChain{
			Handler: func(w *msg.Wrapper) error {
				time.Sleep(time.Second * 5)
				return nil
			},
		}
	})

	m.OnEvent("test_block", func(ctx iface.INodeContext) iface.IChain {
		return &node.DefaultChain{
			Handler: func(w *msg.Wrapper) error {

				val := msg.GetReqCustomField[int](w, "randvalue")
				w.ToBuilder().WithReqCustomFields(msg.Attr{Key: "randvalue", Value: val + 1})
				ctx.Call("mockc", "mockc", "test_block", w)

				return nil
			},
		}
	})

	m.OnEvent("timeout", func(ctx iface.INodeContext) iface.IChain {
		return &node.DefaultChain{
			Handler: func(w *msg.Wrapper) error {
				time.Sleep(time.Second * 5)
				return nil
			},
		}
	})

	m.OnEvent("tcc_succ", func(ctx iface.INodeContext) iface.IChain {
		return &node.DefaultChain{
			Handler: func(w *msg.Wrapper) error {

				transID := msg.GetReqCustomField[string](w, def.KeyTranscationID)
				m.tcc.stateMap[transID] = &tccState{
					originValue:  MockBTccValue,
					currentValue: 111,
					status:       "try",
					createdAt:    time.Now(),
				}

				MockBTccValue = 111
				fmt.Println("succ mock b value", MockBTccValue)
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
