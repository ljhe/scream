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
}
