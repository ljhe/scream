package iface

type IProcess interface {
	Init() error
	Start() error
	WaitClose() error
	Stop() error
	System() ISystem
}
