package iface

type IMsgHandle interface {
	Start() IMsgHandle
	Stop() IMsgHandle
	SetWorkPool(size int)
	Wait()
	PostCb(cb func())
}
