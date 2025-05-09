package iface

import "time"

type Option func(n INetNode)

type INetNode interface {
	Start() INetNode
	Stop()
	GetTyp() string
}

type RuntimeTag interface {
	SetCloseFlag(b bool)
	GetCloseFlag() bool
	SetRunState(b bool)
	GetRunState() bool
}

// TCPSocketOption socket相关设置
type TCPSocketOption interface {
	SetSocketBuff(read, write int, noDelay bool)
	SetSocketDeadline(read, write time.Duration)
}

type ProcessorRPCBundle interface {
	SetMessageProc(v MessageProcessor)
	SetHooker(v HookEvent)
	SetMsgHandle(v IMsgHandle)
	SetMsgRouter(v EventCallBack)
}

type ServerNodeProperty interface {
	SetAddr(a string)
	GetAddr() string
	SetName(s string)
	GetName() string
	SetZone(z int)
	GetZone() int
	SetServerTyp(t int)
	GetServerTyp() int
	SetIndex(i int)
	GetIndex() int
	SetServerNodeProperty()
}

type ContextSet interface {
	SetContextData(key, val interface{})
	GetContextData(key interface{}) (interface{}, bool)
	RawContextData(key interface{}, ptr interface{}) bool
}
