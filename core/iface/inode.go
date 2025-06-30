package iface

import (
	"github.com/ljhe/scream/utils"
)

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

type IOption interface {
	GetMaxMsgLen() int
	SocketReadTimeout(s ISession, callback func())
	SocketWriteTimeout(s ISession, callback func())
	SetOption(opt interface{})
}

type IProcessor interface {
	SetMsgFlow(v IMsgFlow)
	SetHooker(v IHookEvent)
	SetMsgHandle(v IMsgHandle)
	SetMsgRouter(v EventCallBack)
	GetMsgRouter() EventCallBack
}

type INodeProp interface {
	SetAddr(a string)
	GetAddr() string
	SetName(s string)
	GetName() string
	SetServerTyp(t int)
	GetServerTyp() int
	SetIndex(i int)
	GetIndex() int
	SetNodeProp(typ, index int)
}

type IContextSet interface {
	SetContextData(key, val interface{})
	GetContextData(key interface{}) (interface{}, bool)
	RawContextData(key interface{}, ptr interface{}) bool
}

type IDiscover interface {
	// Loader load all node info by ETCD after the node started
	Loader()
	Close()
	GetNodeByKey(key string) *utils.ServerInfo
}
