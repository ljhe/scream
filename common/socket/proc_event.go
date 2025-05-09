package socket

import (
	"github.com/ljhe/scream/common/iface"
)

type RcvMsgEvent struct {
	Sess    iface.ISession
	Message interface{}
	Err     error
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

// DataPacket 收发数据
type DataPacket interface {
	ReadMessage(s iface.ISession) (interface{}, error)
	SendMessage(s iface.ISession, msg interface{}) (err error)
}
