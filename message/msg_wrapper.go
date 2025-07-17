package message

import (
	"context"
	"github.com/google/uuid"
	"github.com/ljhe/scream/lib/warpwaitgroup"
)

type Wrapper struct {
	Req  *Message // The proto-defined Message
	Res  *Message
	Ctx  context.Context
	Err  error
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

// newMessage create new message
func newMessage(uid string) *Message {
	m := &Message{
		Header: &Header{
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

	if wc, ok := ctx.Value(WaitGroupKey{}).(*warpwaitgroup.WrapWaitGroup); ok {
		ctx = context.WithValue(ctx, WaitGroupKey{}, wc)
	} else {
		ctx = context.WithValue(ctx, WaitGroupKey{}, &warpwaitgroup.WrapWaitGroup{})
	}

	return &MsgBuilder{
		wrapper: &Wrapper{
			Ctx: ctx,
			Req: newMessage(uid),
			Res: newMessage(uid),
		},
	}
}

func (b *MsgBuilder) WithReqHeader(h *Header) *MsgBuilder {
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
func (b *MsgBuilder) WithResHeader(h *Header) *MsgBuilder {
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
