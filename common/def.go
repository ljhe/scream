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

type RcvMsgEvent struct {
	Sess    iface.ISession
	Message interface{}
}

func (re *RcvMsgEvent) Session() iface.ISession {
	return re.Sess
}

func (re *RcvMsgEvent) Msg() interface{} {
	return re.Message
}

type SendMsgEvent struct {
	Sess    iface.ISession
	Message interface{}
}

func (se *SendMsgEvent) Session() iface.ISession {
	return se.Sess
}

func (se *SendMsgEvent) Msg() interface{} {
	return se.Message
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
