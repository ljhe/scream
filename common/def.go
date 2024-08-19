package common

import "common/iface"

// IMsgHandle 事件处理队列
type IMsgHandle interface {
	Start() IMsgHandle
	Stop() IMsgHandle
	SetWorkPool(size int)
	Wait()
	PostCb(cb func())
}

type ReceiveMsgEvent struct {
	Sess    iface.ISession
	Message interface{}
}

func (r *ReceiveMsgEvent) Session() iface.ISession {
	return r.Sess
}

func (r *ReceiveMsgEvent) Msg() interface{} {
	return r.Message
}

type EventHook interface {
	InEvent(iv iface.IProcEvent) iface.IProcEvent  // 接收事件
	OutEvent(ov iface.IProcEvent) iface.IProcEvent // 发送事件
}

// MessageProcessor 消息处理
type MessageProcessor interface {
	OnRcvMsg(s iface.ISession) (interface{}, uint32, error)
	OnSendMsg(s iface.ISession, msg interface{}) error
}
