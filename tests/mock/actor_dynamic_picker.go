package mock

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ljhe/scream/core"
	"github.com/ljhe/scream/core/actor"
	"github.com/ljhe/scream/def"
	"github.com/ljhe/scream/router/msg"
	"golang.org/x/time/rate"
)

type dynamicPickerActor struct {
	*actor.Runtime
	limiter *rate.Limiter
}

func NewDynamicPickerActor(p core.IActorBuilder) core.IActor {
	return &dynamicPickerActor{
		Runtime: &actor.Runtime{Id: p.GetID(), Ty: "MockDynamicPicker", Sys: p.GetSystem()},
		limiter: rate.NewLimiter(rate.Every(time.Second/200), 1), // 允许每秒10次调用
	}
}

func (a *dynamicPickerActor) Init(ctx context.Context) {
	a.Runtime.Init(ctx)

	a.OnEvent("MockDynamicPick", func(ctx core.ActorContext) core.IChain {
		return &actor.DefaultChain{

			Handler: func(mw *msg.Wrapper) error {

				// 使用限流器
				if err := a.limiter.Wait(mw.Ctx); err != nil {
					return err
				}

				actor_ty := msg.GetReqCustomField[string](mw, def.KeyActorTy)

				// Select a node with low weight and relatively fewer registered actors of this type
				nodeaddr, err := ctx.AddressBook().GetLowWeightNodeForActor(mw.Ctx, actor_ty)
				if err != nil {
					return err
				}

				// rename
				msgbuild := mw.ToBuilder().WithReqCustomFields(def.ActorID(nodeaddr.Node + "_" + actor_ty + "_" + uuid.NewString()))

				// dispatcher to picker node
				return ctx.Call(nodeaddr.Node+"_"+"MockDynamicRegister", "MockDynamicRegister", "MockDynamicRegister", msgbuild.Build())
			},
		}
	})
}
