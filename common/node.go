package common

import (
	"time"
)

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
}

// TCPSocketOption option.go
type TCPSocketOption interface {
	SetSocketBuff(read, write int, noDelay bool)
	SetSocketDeadline(read, write time.Duration)
}

type ProcessorRPCBundle interface {
	SetMessageProc(v MessageProcessor)
	SetHooker(v EventHook)
	SetMsgHandle(v IMsgHandle)
	SetMsgRouter(v EventCallBack)
}

type ContextSet interface {
	SetContextData(key, val interface{})
	GetContextData(key interface{}) (interface{}, bool)
	RawContextData(key interface{}, ptr interface{}) bool
}
