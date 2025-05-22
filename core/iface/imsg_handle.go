package iface

type IMsgHandle interface {
	Start() IMsgHandle
	Stop() IMsgHandle
	Wait()
	PostCb(cb func())
}
