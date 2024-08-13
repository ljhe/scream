package socket

import (
	"common"
	"common/iface"
	"fmt"
)

type NetProcessorRPC struct {
	MsgHandle common.IMsgHandle
}

func (n *NetProcessorRPC) ProcEvent(e iface.IProcEvent) {
	n.MsgHandle.PostCb(func() {
		fmt.Println("这里是测试数据")
	})
}
