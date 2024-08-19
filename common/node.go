package common

type ServerNodeProperty interface {
	GetAddr() string
	SetAddr(s string)
}

type ProcessorRPCBundle interface {
	SetMessageProc(v MessageProcessor)
	SetHooker(v EventHook)
	SetMsgHandle(v IMsgHandle)
}
