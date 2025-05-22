package iface

type IMsgFlow interface {
	OnRcvMsg(s ISession) (interface{}, error)
	OnSendMsg(s ISession, msg interface{}) error
}

// IDataPacket 收发数据
type IDataPacket interface {
	ReadMessage(s ISession) (interface{}, error)
	SendMessage(s ISession, msg interface{}) (err error)
}
