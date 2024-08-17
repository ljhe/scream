package common

type ServerNodeProperty interface {
	GetAddr() string
	SetAddr(s string)
}

type ProcessorRPCBundle interface {
	SetHooker(v EventHook)
	SetMsgHandle(v IMsgHandle)
}
