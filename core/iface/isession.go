package iface

type ISessionCommon interface {
	Conn() interface{} // 获得conn
	TransmitChild(sessionId uint64, data interface{})
	DelChild(sessionId uint64)
	HeartBeat(msg interface{}) // 心跳检测
}

type ISession interface {
	ISessionCommon
	Node() INetNode
	GetProcessor() IProcessor
	Send(msg interface{})
	Close()
	SetId(id uint64)
	GetId() uint64
	IncRcvPingNum(inc int)
	RcvPingNum() int
	Start()
	RunRcv()
	RunSend()
	ConnClose()
}

type ISessionExtension interface {
	ISessionCommon
	SetConn(c interface{})
	CloseEvent(err error)
}

type ISessionChild interface {
	Start()
	Stop()
	Rcv(msg interface{})
	GetSessionId() uint64
}

type ISessionManager interface {
	Add(s ISession)
	Get(sessionId uint64) (ISession, bool)
	Remove(s ISession)
	SetUuidCreateKey(genKey int)
	CloseAllSession()
}
