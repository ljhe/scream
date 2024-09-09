package iface

import "net"

type ISession interface {
	SetConn(c net.Conn)
	GetConn() net.Conn
	Node() INetNode
	Send(msg interface{})
	Close()
	SetId(id uint64)
	GetId() uint64
	HeartBeat(msg interface{})
	IncRcvPingNum(inc int)
	RcvPingNum() int
}
