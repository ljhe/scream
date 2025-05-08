package socket

import (
	"fmt"
	"github.com/ljhe/scream/common"
	"github.com/ljhe/scream/common/iface"
	"log"
	"reflect"
	"time"
)

type NetProcessorRPC struct {
	MsgRPC    common.MessageProcessor // 根据不同对象来处理消息读写的加解密
	Hooker    common.EventHook        // 不进入主消息队列 直接操作
	MsgHandle common.IMsgHandle
	MsgRouter common.EventCallBack
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

func (n *NetProcessorRPC) SetMsgRouter(msgr common.EventCallBack) {
	n.MsgRouter = msgr
}

func (n *NetProcessorRPC) GetRPC() *NetProcessorRPC {
	return n
}

func (n *NetProcessorRPC) ProcEvent(e iface.IProcEvent) {
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

type TCPMessageProcessor struct {
}

func (tp *TCPMessageProcessor) OnRcvMsg(s iface.ISession) (msg interface{}, err error) {
	opt := s.Node().(Option)
	opt.SocketReadTimeout(s, func() {
		p := TcpDataPacket{}
		msg, err = p.ReadMessage(s)
	})
	return
}

func (tp *TCPMessageProcessor) OnSendMsg(s iface.ISession, msg interface{}) (err error) {
	opt := s.Node().(Option)
	opt.SocketWriteTimeout(s, func() {
		p := TcpDataPacket{}
		err = p.SendMessage(s, msg)
	})
	return err
}

type WSMessageProcessor struct {
}

func (tp *WSMessageProcessor) OnRcvMsg(s iface.ISession) (msg interface{}, err error) {
	opt := s.Node().(Option)
	opt.SocketReadTimeout(s, func() {
		p := WsDataPacket{}
		msg, err = p.ReadMessage(s)
	})
	return
}

func (tp *WSMessageProcessor) OnSendMsg(s iface.ISession, msg interface{}) (err error) {
	opt := s.Node().(Option)
	opt.SocketWriteTimeout(s, func() {
		p := WsDataPacket{}
		err = p.SendMessage(s, msg)
	})
	return err
}
