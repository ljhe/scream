package socket

import (
	"github.com/ljhe/scream/core/iface"
)

type RcvProcEvent struct {
	Sess    iface.ISession
	Message interface{}
	Err     error
}

func (re *RcvProcEvent) Session() iface.ISession {
	return re.Sess
}

func (re *RcvProcEvent) Msg() interface{} {
	return re.Message
}

type SendProcEvent struct {
	Sess    iface.ISession
	Message interface{}
}

func (se *SendProcEvent) Session() iface.ISession {
	return se.Sess
}

func (se *SendProcEvent) Msg() interface{} {
	return se.Message
}
