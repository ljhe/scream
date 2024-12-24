package iface

type Option func(n INetNode)

type INetNode interface {
	Start() INetNode
	Stop()
	GetTyp() string
}
