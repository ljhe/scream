package iface

import "net"

type ISession interface {
	SetConn(c net.Conn)
	GetConn() net.Conn
	Node() INetNode
	Send(msg interface{})
	HeartBeat(msg interface{})
}
