package socket

import (
	"fmt"
	"github.com/ljhe/scream/common/iface"
	"log"
	"reflect"
	"time"
)

type NetProcessorRPC struct {
	MsgRPC    iface.MessageProcessor // 根据不同对象来处理消息读写的加解密
	Hooker    iface.HookEvent        // 不进入主消息队列 直接操作
	MsgHandle iface.IMsgHandle
	MsgRouter iface.EventCallBack
	count     int // 记录当前队列中消息数量
}

func (n *NetProcessorRPC) SetMessageProc(v iface.MessageProcessor) {
	n.MsgRPC = v
}

func (n *NetProcessorRPC) SetHooker(v iface.HookEvent) {
	n.Hooker = v
}

func (n *NetProcessorRPC) SetMsgHandle(v iface.IMsgHandle) {
	n.MsgHandle = v
}

func (n *NetProcessorRPC) SetMsgRouter(msgr iface.EventCallBack) {
	n.MsgRouter = msgr
}

func (n *NetProcessorRPC) GetRPC() *NetProcessorRPC {
	return n
}

func (n *NetProcessorRPC) ProcEvent(e iface.IProcEvent) {
	n.count++
	if n.Hooker != nil {
		e = n.Hooker.InEvent(e)
	}
	if e != nil {
		if n.MsgHandle != nil {
			n.MsgHandle.PostCb(func() {
				start := time.Now()
				n.MsgRouter(e)
				duration := time.Since(start)
				log.Printf("%+v 方法 耗时: %s (%dμs / %dns)\n", reflect.TypeOf(e.Msg()), duration, duration.Microseconds(), duration.Nanoseconds())
			})
		}
	}
	fmt.Printf("proc event msg count:%d \n", n.count)
}

func (n *NetProcessorRPC) ReadMsg(s iface.ISession) (interface{}, error) {
	if n.MsgRPC != nil {
		return n.MsgRPC.OnRcvMsg(s)
	}
	return nil, fmt.Errorf("msg rpc is nil")
}

func (n *NetProcessorRPC) SendMsg(e iface.IProcEvent) error {
	if n.Hooker != nil {
		e = n.Hooker.OutEvent(e)
	}
	if n.MsgRPC != nil {
		return n.MsgRPC.OnSendMsg(e.Session(), e.Msg())
	}
	return nil
}
