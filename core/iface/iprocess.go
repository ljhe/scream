package iface

type IProcess interface {
	ID() string
	PID() int
	GetHost() string

	Start()
	WaitExitSignal()
	Stop()

	RegisterNode(node INetNode) error
	GetNode(id string) INetNode
	GetAllNodes() []INetNode
}
