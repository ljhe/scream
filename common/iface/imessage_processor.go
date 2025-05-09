package iface

type MessageProcessor interface {
	OnRcvMsg(s ISession) (interface{}, error)
	OnSendMsg(s ISession, msg interface{}) error
}
