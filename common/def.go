package common

// IMsgHandle 事件处理队列
type IMsgHandle interface {
	Start() IMsgHandle
	Stop() IMsgHandle
	SetWorkPool(size int)
	Wait()
	PostCb(cb func())
}

type ReceiveMsgEvent struct {
	Message interface{}
}

func (r *ReceiveMsgEvent) Msg() interface{} {
	return r.Message
}
