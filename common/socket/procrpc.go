package socket

import (
	"common"
	"common/iface"
	"fmt"
	"github.com/gorilla/websocket"
	"net"
)

type NetProcessorRPC struct {
	MsgRPC    common.MessageProcessor // 根据不同对象来处理消息读写的加解密
	Hooker    common.EventHook        // 不进入主消息队列 直接操作
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
	opt.SocketReadTimeout(s.Raw().(net.Conn), func() {
		p := TcpDataPacket{}
		msg, err = p.ReadMessage(s)
	})
	return
}

func (tp *TCPMessageProcessor) OnSendMsg(s iface.ISession, msg interface{}) (err error) {
	opt := s.Node().(Option)
	opt.SocketWriteTimeout(s.Raw().(net.Conn), func() {
		p := TcpDataPacket{}
		err = p.SendMessage(s, msg)
	})
	return err
}

type WSMessageProcessor struct {
}

func (tp *WSMessageProcessor) OnRcvMsg(s iface.ISession) (msg interface{}, err error) {
	opt := s.Node().(Option)
	opt.WSReadTimeout(s.Raw().(*websocket.Conn), func() {
		p := WsDataPacket{}
		msg, err = p.ReadMessage(s)
	})
	return
}

func (tp *WSMessageProcessor) OnSendMsg(s iface.ISession, msg interface{}) (err error) {
	opt := s.Node().(Option)
	opt.WSWriteTimeout(s.Raw().(*websocket.Conn), func() {
		p := WsDataPacket{}
		err = p.SendMessage(s, msg)
	})
	return err
}
