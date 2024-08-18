package socket

import (
	"common"
	"common/iface"
	"fmt"
)

type NetProcessorRPC struct {
	Hooker    common.EventHook // 不进入主消息队列 直接操作
	MsgHandle common.IMsgHandle
}

func (n *NetProcessorRPC) SetHooker(v common.EventHook) {
	n.Hooker = v
}

func (n *NetProcessorRPC) SetMsgHandle(v common.IMsgHandle) {
	n.MsgHandle = v
}

func (n *NetProcessorRPC) ProcEvent(e iface.IProcEvent) {
	if n.Hooker != nil {
		e = n.Hooker.InEvent(e)
	}
	if e != nil {
		n.MsgHandle.PostCb(func() {
			fmt.Println("这里是测试数据")
		})
	}
}

func (n *NetProcessorRPC) SendMsg(e iface.IProcEvent) error {
	if n.Hooker != nil {
		e = n.Hooker.OutEvent(e)
	}
	return nil
}
