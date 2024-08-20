package socket

import (
	"common"
	"common/iface"
	"fmt"
	"io"
	"log"
	"net"
)

type NetProcessorRPC struct {
	MsgRPC    common.MessageProcessor
	Hooker    common.EventHook // 不进入主消息队列 直接操作
	MsgHandle common.IMsgHandle
}

func (n *NetProcessorRPC) SetMessageProc(v common.MessageProcessor) {
	n.MsgRPC = v
}

func (n *NetProcessorRPC) SetHooker(v common.EventHook) {
	n.Hooker = v
}

func (n *NetProcessorRPC) SetMsgHandle(v common.IMsgHandle) {
	n.MsgHandle = v
}

func (n *NetProcessorRPC) GetRPC() *NetProcessorRPC {
	return n
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
	if n.MsgRPC != nil {
		return n.MsgRPC.OnSendMsg(e.Session(), e.Msg())
	}
	return nil
}

type TCPMessageProcessor struct {
}

func (tp *TCPMessageProcessor) OnRcvMsg(s iface.ISession) (msg interface{}, msgSeqId uint32, err error) {
	return nil, 0, err
}

func (tp *TCPMessageProcessor) OnSendMsg(s iface.ISession, msg interface{}) (err error) {
	w, ok := s.GetConn().(io.Writer)
	if !ok || w == nil {
		log.Println("conn is not io.Writer")
		return fmt.Errorf("conn is not io.Writer")
	}
	opt := s.Node().(Option)
	opt.SocketWriteTimeout(w.(net.Conn), func() {
		err = SendMessage(w, msg)
	})
	return err
}
