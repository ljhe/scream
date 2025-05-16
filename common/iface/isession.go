package iface

type ISession interface {
	Raw() interface{} // 获得conn
	Node() INetNode
	Send(msg interface{})
	Close()
	SetId(id uint64)
	GetId() uint64
	HeartBeat(msg interface{})
	IncRcvPingNum(inc int)
	RcvPingNum() int
	SetSessionChild(sessionId uint64, data interface{})
	Start()
}

type ISessionChild interface {
	Start(sessionId uint64)
	Stop(sessionId uint64)
	Rcv(msg interface{})
}

type ISessionManager interface {
	Add(s ISession)
	Get(sessionId uint64) (ISession, bool)
	Remove(s ISession)
	SetUuidCreateKey(genKey int)
	CloseAllSession()
}
