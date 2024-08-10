package iface

type INetNode interface {
	Start() INetNode
	Stop()
	GetTyp() string
}
