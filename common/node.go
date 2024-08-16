package common

type ServerNodeProperty interface {
	GetAddr() string
	SetAddr(s string)
}

type ProcessorRPCBundle interface {
	SetMsgHandle(v IMsgHandle)
}
