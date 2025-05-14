package iface

type IMessageProcessor interface {
	OnRcvMsg(s ISession) (interface{}, error)
	OnSendMsg(s ISession, msg interface{}) error
}
