package socket

import (
	"fmt"
	"github.com/ljhe/scream/common/iface"
	"log"
	"reflect"
	"time"
)

type Processor struct {
	MsgProc   iface.IMessageProcessor // 根据不同对象来处理消息读写的加解密
	Hooker    iface.IHookEvent        // 不进入主消息队列 直接操作
	MsgHandle iface.IMsgHandle
	MsgRouter iface.EventCallBack
}

func (n *Processor) SetMessageProc(v iface.IMessageProcessor) {
	n.MsgProc = v
}

func (n *Processor) SetHooker(v iface.IHookEvent) {
	n.Hooker = v
}

func (n *Processor) SetMsgHandle(v iface.IMsgHandle) {
	n.MsgHandle = v
}

func (n *Processor) SetMsgRouter(msgr iface.EventCallBack) {
	n.MsgRouter = msgr
}

func (n *Processor) GetRPC() *Processor {
	return n
}

func (n *Processor) ProcEvent(e iface.IProcEvent) {
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
}

func (n *Processor) ReadMsg(s iface.ISession) (interface{}, error) {
	if n.MsgProc != nil {
		return n.MsgProc.OnRcvMsg(s)
	}
	return nil, fmt.Errorf("msg rpc is nil")
}

func (n *Processor) SendMsg(e iface.IProcEvent) error {
	if n.Hooker != nil {
		e = n.Hooker.OutEvent(e)
	}
	if n.MsgProc != nil {
		return n.MsgProc.OnSendMsg(e.Session(), e.Msg())
	}
	return nil
}
