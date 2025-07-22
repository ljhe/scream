package mock

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/core/node"
	"github.com/ljhe/scream/def"
	"github.com/ljhe/scream/msg"
	"sync/atomic"
)

type mockNodeA struct {
	*node.Node
}

var RecenterCalcValue int32

// 添加一个计数器
var ReceivedMessageCount int64

func NewMockA(p iface.INodeBuilder) iface.INode {
	return &mockNodeA{
		Node: &node.Node{Id: p.GetID(), Ty: p.GetType(), Sys: p.GetSystem()},
	}
}

func (m *mockNodeA) Init(ctx context.Context) {
	m.Node.Init(ctx)

	m.OnEvent("reenter", func(ctx iface.INodeContext) iface.IChain {
		return &node.DefaultChain{
			Handler: func(w *msg.Wrapper) error {

				calculateVal := 2
				w.ToBuilder().WithReqCustomFields(msg.Attr{Key: "calculateVal", Value: calculateVal})

				future := ctx.ReenterCall("mockb", "mockb", "clac", w)
				future.Then(func(w *msg.Wrapper) {
					val := msg.GetResCustomField[int](w, "calculateVal")
					atomic.CompareAndSwapInt32(&RecenterCalcValue, 0, int32(val*2))
				})

				return nil
			},
		}
	})

	m.OnEvent("call_benchmark", func(ctx iface.INodeContext) iface.IChain {
		return &node.DefaultChain{
			Handler: func(w *msg.Wrapper) error {
				ctx.Call("mockb", "mockb", "call_benchmark", w)
				return nil
			},
		}
	})

	m.OnEvent("timeout", func(ctx iface.INodeContext) iface.IChain {
		return &node.DefaultChain{
			Handler: func(w *msg.Wrapper) error {

				future := ctx.ReenterCall("mockb", "mockb", "timeout", w)
				future.Then(func(fw *msg.Wrapper) {

					if fw.Err != nil {
						w.Err = fw.Err
					}
				})

				return nil
			},
		}
	})

	m.OnEvent("chain", func(ctx iface.INodeContext) iface.IChain {
		return &node.DefaultChain{
			Handler: func(w *msg.Wrapper) error {

				calculateVal := 2
				w.ToBuilder().WithReqCustomFields(msg.Attr{Key: "calculateVal", Value: calculateVal})

				future := ctx.ReenterCall("mockb", "mockb", "clac", w)
				future.Then(func(w *msg.Wrapper) {
					val := msg.GetResCustomField[int](w, "calculateVal")
					atomic.CompareAndSwapInt32(&RecenterCalcValue, 0, int32(val*2))
				}).Then(func(w *msg.Wrapper) {
					atomic.AddInt32(&RecenterCalcValue, 10)
				})

				return nil
			},
		}
	})

	m.OnEvent("test_block", func(ctx iface.INodeContext) iface.IChain {
		return &node.DefaultChain{
			Handler: func(w *msg.Wrapper) error {

				val := msg.GetReqCustomField[int](w, "randvalue")
				w.ToBuilder().WithReqCustomFields(msg.Attr{Key: "randvalue", Value: val + 1})
				ctx.Call("mockb", "mockb", "test_block", w)

				return nil
			},
		}
	})

	m.OnEvent("tcc_succ", func(ctx iface.INodeContext) iface.IChain {
		return &node.DefaultChain{
			Handler: func(w *msg.Wrapper) error {

				transactionID := uuid.New().String()

				bmsg := msg.NewBuilder(w.Ctx).WithReqCustomFields(def.TransactionID(transactionID))
				cmsg := msg.NewBuilder(w.Ctx).WithReqCustomFields(def.TransactionID(transactionID))

				bsucc := ctx.Call("mockb", "mockb", "tcc_succ", bmsg.Build())
				csucc := ctx.Call("mockc", "mockc", "tcc_succ", cmsg.Build())

				var err error

				if bsucc == nil && csucc == nil { // succ

					bconfirmmsg := msg.NewBuilder(w.Ctx).WithReqCustomFields(def.TransactionID(transactionID)).Build()
					err = ctx.Call("mockb", "mockb", "tcc_confirm", bconfirmmsg)
					if err != nil {
						/*
							err = ctx.Pub("mockb_tcc_confirm", bconfirmmsg.Req)
							if err != nil {
								fmt.Println("???")
							}
						*/
					}

					cconfirmmsg := msg.NewBuilder(w.Ctx).WithReqCustomFields(def.TransactionID(transactionID)).Build()
					err = ctx.Call("mockc", "mockc", "tcc_confirm", cconfirmmsg)
					if err != nil {
						/*
							err = ctx.Pub("mockc_tcc_confirm", cconfirmmsg.Req)
							if err != nil {
								fmt.Println("???")
							}
						*/
					}
				} else {
					fmt.Println("tcc call err", "b", bsucc, "c", csucc)
				}

				fmt.Println("mock a tcc_succ end")
				return nil
			},
		}
	})

	m.Sub(m.Id, "offline_msg", func(ctx iface.INodeContext) iface.IChain {
		return &node.DefaultChain{
			Handler: func(w *msg.Wrapper) error {
				atomic.AddInt64(&ReceivedMessageCount, 1)
				return nil
			},
		}
	})
}
