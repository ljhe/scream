package mock

import (
	"context"
	"github.com/ljhe/scream/core/iface"
	"github.com/ljhe/scream/core/node"
	"github.com/ljhe/scream/msg"
	"golang.org/x/time/rate"
	"time"
)

type dynamicPickerActor struct {
	*node.Node
	limiter *rate.Limiter
}

func NewDynamicPickerActor(p iface.INodeBuilder) iface.INode {
	return &dynamicPickerActor{
		Node:    &node.Node{Id: p.GetID(), Ty: "MockDynamicPicker", Sys: p.GetSystem()},
		limiter: rate.NewLimiter(rate.Every(time.Second/200), 1), // 允许每秒10次调用
	}
}

func (a *dynamicPickerActor) Init(ctx context.Context) {
	a.Node.Init(ctx)

	a.OnEvent("MockDynamicPick", func(ctx iface.INodeContext) iface.IChain {
		return &node.DefaultChain{

			Handler: func(mw *msg.Wrapper) error {

				// 使用限流器
				if err := a.limiter.Wait(mw.Ctx); err != nil {
					return err
				}

				//actor_ty := msg.GetReqCustomField[string](mw, def.KeyNodeTy)

				//// Select a node with low weight and relatively fewer registered actors of this type
				//nodeaddr, err := ctx.AddressBook().GetLowWeightNodeForActor(mw.Ctx, actor_ty)
				//if err != nil {
				//	return err
				//}
				//
				//// rename
				//msgbuild := mw.ToBuilder().WithReqCustomFields(def.NodeID(nodeaddr.Node + "_" + actor_ty + "_" + uuid.NewString()))
				//
				//// dispatcher to picker node
				//return ctx.Call(nodeaddr.Node+"_"+"MockDynamicRegister", "MockDynamicRegister", "MockDynamicRegister", msgbuild.Build())
				return nil
			},
		}
	})
}
