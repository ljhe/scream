package iface

type IProcess interface {
	Init() error
	Start() error
	WaitClose() error
	Stop() error
	ID() string
	System() ISystem
}
