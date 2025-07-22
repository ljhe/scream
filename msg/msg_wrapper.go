package msg

import (
	"context"
	"github.com/google/uuid"
	"github.com/ljhe/scream/lib/warpwaitgroup"
	"github.com/ljhe/scream/msg/router"
)

// WrapperParm nsq config
type WrapperParm struct {
	CustomObjSerialize ICustomSerialize // default msg pack
	CustomMapSerialize ICustomSerialize // default json
}

type Wrapper struct {
	Req *router.Message // The proto-defined Message
	Res *router.Message
	Ctx context.Context
	Err error

	parm WrapperParm
	Done chan struct{} // Used for synchronization
}

func (mw *Wrapper) ToBuilder() *MsgBuilder {
	if mw == nil {
		return NewBuilder(context.Background())
	}

	return &MsgBuilder{
		wrapper: mw,
	}
}

func (mw *Wrapper) GetWg() *warpwaitgroup.WrapWaitGroup {
	if wc, ok := mw.Ctx.Value(WaitGroupKey{}).(*warpwaitgroup.WrapWaitGroup); ok {
		return wc
	}
	return nil
}

// newMessage create new msg
func newMessage(uid string) *router.Message {
	m := &router.Message{
		Header: &router.Header{
			ID: uid,
		},
	}
	return m
}

type WaitGroupKey struct{}

// MsgBuilder used to build MsgWrapper
type MsgBuilder struct {
	wrapper *Wrapper
}

func NewBuilder(ctx context.Context) *MsgBuilder {
	uid := uuid.NewString()

	parm := WrapperParm{
		CustomObjSerialize: &CustomObjectSerialize{},
		CustomMapSerialize: &CustomJsonSerialize{},
	}

	if wc, ok := ctx.Value(WaitGroupKey{}).(*warpwaitgroup.WrapWaitGroup); ok {
		ctx = context.WithValue(ctx, WaitGroupKey{}, wc)
	} else {
		ctx = context.WithValue(ctx, WaitGroupKey{}, &warpwaitgroup.WrapWaitGroup{})
	}

	return &MsgBuilder{
		wrapper: &Wrapper{
			parm: parm,
			Ctx:  ctx,
			Req:  newMessage(uid),
			Res:  newMessage(uid),
		},
	}
}

func Swap(mw *Wrapper) *Wrapper {

	ctx := context.WithValue(mw.Ctx, WaitGroupKey{}, &warpwaitgroup.WrapWaitGroup{})

	return &Wrapper{
		Ctx: ctx,
		// 交换 Req 和 Res
		parm: mw.parm,
		Req:  mw.Req,
		Res:  mw.Res,
		Done: make(chan struct{}),
	}
}

func (b *MsgBuilder) WithReqHeader(h *router.Header) *MsgBuilder {
	if b.wrapper.Req.Header != nil && h != nil {
		// Copy fields from the input header to existing header
		b.wrapper.Req.Header.ID = h.ID
		b.wrapper.Req.Header.Event = h.Event
		b.wrapper.Req.Header.OrgActorID = h.OrgActorID
		b.wrapper.Req.Header.OrgActorType = h.OrgActorType
		b.wrapper.Req.Header.Token = h.Token
		b.wrapper.Req.Header.PrevActorType = h.PrevActorType
		b.wrapper.Req.Header.TargetActorID = h.TargetActorID
		b.wrapper.Req.Header.TargetActorType = h.TargetActorType
		b.wrapper.Req.Header.Custom = h.Custom
	} else {
		// If either header is nil, directly set the header
		b.wrapper.Req.Header = h
	}
	return b
}

func (b *MsgBuilder) WithReqBody(byt []byte) *MsgBuilder {
	b.wrapper.Req.Body = byt
	return b
}

// WithResHeader set res header
func (b *MsgBuilder) WithResHeader(h *router.Header) *MsgBuilder {
	b.wrapper.Res.Header = h
	return b
}

func (b *MsgBuilder) WithResBody(byt []byte) *MsgBuilder {
	b.wrapper.Res.Body = byt
	return b
}

// Build msg wrapper
func (b *MsgBuilder) Build() *Wrapper {
	return b.wrapper
}
