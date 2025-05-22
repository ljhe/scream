package iface

import "time"

type INetNode interface {
	Start() INetNode
	Stop()
	GetTyp() string
}

type IRuntimeTag interface {
	SetCloseFlag(b bool)
	GetCloseFlag() bool
	SetRunState(b bool)
	GetRunState() bool
}

// ITCPSocketOption socket相关设置
type ITCPSocketOption interface {
	SetSocketBuff(read, write int, noDelay bool)
	SetSocketDeadline(read, write time.Duration)
}

type IProcessor interface {
	SetMessageProc(v IMessageProcessor)
	SetHooker(v IHookEvent)
	SetMsgHandle(v IMsgHandle)
	SetMsgRouter(v EventCallBack)
	GetMsgRouter() EventCallBack
}

type IServerNodeProperty interface {
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

type IContextSet interface {
	SetContextData(key, val interface{})
	GetContextData(key interface{}) (interface{}, bool)
	RawContextData(key interface{}, ptr interface{}) bool
}
