package iface

type ISession interface {
	Conn() interface{} // 获得conn
	Node() INetNode
	Send(msg interface{})
	Close()
	SetId(id uint64)
	GetId() uint64
	HeartBeat(msg interface{})
	IncRcvPingNum(inc int)
	RcvPingNum() int
	SetSessionChild(sessionId uint64, data interface{})
	DelSessionChild(sessionId uint64)
	Start()
}

type ISessionSpecific interface {
	Conn() interface{}
	SetConn(c interface{})
	RunRcv()
	SetSessionChild(sessionId uint64, data interface{})
	DelSessionChild(sessionId uint64)
	HeartBeat(msg interface{})
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
