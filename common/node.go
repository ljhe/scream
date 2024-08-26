package common

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

type ProcessorRPCBundle interface {
	SetMessageProc(v MessageProcessor)
	SetHooker(v EventHook)
	SetMsgHandle(v IMsgHandle)
}
